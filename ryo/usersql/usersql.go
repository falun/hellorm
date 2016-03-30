package usersql

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/falun/hellorm/ryo"
	"github.com/falun/hellorm/ryo/orgsql"
	"github.com/falun/hellorm/ryo/permsql"
	"github.com/falun/hellorm/ryo/userpermsql"
)

type Column struct {
	Name string
	Ex   func(ryo.User) interface{}
}

var (
	ColKey       Column = Column{"key", func(u ryo.User) interface{} { return u.Key }}
	ColLogin            = Column{"login", func(u ryo.User) interface{} { return u.Login }}
	ColEmail            = Column{"email", func(u ryo.User) interface{} { return u.Email }}
	ColDeletedAt        = Column{"deleted_ts", func(u ryo.User) interface{} { return u.DeletedAt }}
	ColOrgKey           = Column{"org_key", func(u ryo.User) interface{} { return u.OrgKey }}
	ColChecksum         = Column{"checksum", func(u ryo.User) interface{} { return u.Checksum }}
)

type User struct {
	Db              *ryo.DB
	Org             *orgsql.Org
	Permission      *permsql.Permission
	UserPermissions *userpermsql.UserPermission
}

var (
	TableName = "users"

	DropTable   = fmt.Sprintf(`DROP TABLE %q`, TableName)
	CreateTable = fmt.Sprintf(`
    CREATE TABLE %q (
      key character varying(128) NOT NULL,
      login character varying(64) NOT NULL DEFAULT ''::character varying,
      email character varying(128) NOT NULL DEFAULT ''::character varying,
      deleted_ts timestamp with time zone,
      org_key character varying(128) NOT NULL,
      checksum character varying(128) NOT NULL,

      CONSTRAINT user_pkey PRIMARY KEY (key),
      CONSTRAINT user_email_key UNIQUE (email),
      CONSTRAINT user_key_login_key UNIQUE (key, login),
      CONSTRAINT users_org_key_orgs_key_foreign FOREIGN KEY (org_key)
        REFERENCES orgs(key) MATCH SIMPLE
        ON UPDATE CASCADE ON DELETE RESTRICT,
      CONSTRAINT users_email_key UNIQUE (email)
    )`, TableName)

	ReadByKey = fmt.Sprintf(`
    SELECT
		  "key", "login", "email", "deleted_ts", "org_key", "checksum"
    FROM %q
    WHERE key = $1`, TableName)
	Insert = fmt.Sprintf(`
    INSERT INTO %q ("key", "login", "email", "deleted_ts", "org_key", "checksum")
    VALUES($1, $2, $3, $4, $5, $6)`, TableName)

	UpdateBase            = fmt.Sprintf(`UPDATE %q SET (%%s) = (%%s) WHERE %q."key" = $1`, TableName, TableName)
	ConditionalUpdateBase = fmt.Sprintf(
		`UPDATE %q SET (%%s) = (%%s) WHERE "checksum" = $2 AND "key" = $1`, TableName)
)

func uniqCols(c ...Column) []Column {
	var cmap map[string]Column
	for _, c := range c {
		cmap[c.Name] = c
	}

	r := make([]Column, 0, len(cmap))
	for _, v := range cmap {
		r = append(r, v)
	}

	return r
}

func mkUpdateSql(base string, u ryo.User, startCounting int, cols ...Column) (string, []interface{}) {
	// track what we've seen to ensure we don't double count something if a
	// client passes in multiple instances of the same column
	seen := map[string]bool{}
	col_names := []string{}
	sql_params := []string{}
	values := []interface{}{}

	for i, e := range cols {
		if seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		col_names = append(col_names, fmt.Sprintf("%q", e.Name))
		sql_params = append(sql_params, "$"+strconv.Itoa(i+startCounting))
		values = append(values, e.Ex(u))
	}

	return fmt.Sprintf(
		base,
		strings.Join(col_names, ","),
		strings.Join(sql_params, ","),
	), values
}

func partialUpdate(db *ryo.DB, user ryo.User, cols ...Column) (int64, error) {
	query, updateValues := mkUpdateSql(UpdateBase, user, 2, cols...)
	values := []interface{}{user.Key}
	values = append(values, updateValues...)

	r, e := db.Exec(query, values...)
	n, e2 := r.RowsAffected()

	if e != nil {
		return n, e
	}
	if e2 != nil {
		return n, e2
	}

	return n, nil
}

func partialUpdateCheckCS(db *ryo.DB, user ryo.User, prevCS string, cols ...Column) (int64, error) {
	query, updateValues := mkUpdateSql(ConditionalUpdateBase, user, 3, cols...)
	values := []interface{}{user.Key, prevCS}
	values = append(values, updateValues...)
	r, e := db.Exec(query, values...)
	n, e2 := r.RowsAffected()

	if e != nil {
		return n, e
	}
	if e2 != nil {
		return n, e2
	}

	return n, nil
}

func insert(db *ryo.DB, user ryo.User, savePerms bool) error {
	_, e := db.Exec(
		Insert,
		user.Key, user.Login, user.Email, user.DeletedAt, user.OrgKey, user.Checksum)
	return e
}

func (u *User) Insert(newUser ryo.User) error {
	return insert(u.Db, newUser, false)
}

type UpdateOpts struct {
	VerifyChecksum   bool
	PreviousChecksum string
}

func (u *User) PartialUpdate(user ryo.User, opts *UpdateOpts, updateCols ...Column) (int64, error) {
	if opts == nil || !opts.VerifyChecksum {
		return partialUpdate(u.Db, user, updateCols...)
	}

	return partialUpdateCheckCS(u.Db, user, opts.PreviousChecksum, updateCols...)
}

func (u *User) Update(user ryo.User, oldCS string) (int64, error) {
	allFields := []Column{
		ColKey, ColLogin, ColEmail, ColDeletedAt, ColOrgKey, ColChecksum}
	return u.PartialUpdate(user, &UpdateOpts{true, oldCS}, allFields...)
}

type GetOpts struct {
	LoadPermissions bool
	LoadOrg         bool
}

func (u *User) Get(userKey string, opts *GetOpts) (ryo.User, error) {
	usr := ryo.User{DeletedAt: new(time.Time)}

	err := u.Db.QueryRow(ReadByKey, userKey).Scan(
		&usr.Key, &usr.Login, &usr.Email, &usr.DeletedAt, &usr.OrgKey, &usr.Checksum)

	if err != nil {
		return ryo.User{}, err
	}

	if opts != nil {
		if opts.LoadOrg {
			o, err := u.Org.Get(usr.OrgKey)
			if err != nil {
				return ryo.User{}, err
			}
			usr.Org = &o
		}

		if opts.LoadPermissions {
			ps, err := u.Permission.GrantedTo(usr.Key)
			if err != nil {
				return ryo.User{}, err
			}
			usr.Permissions = &ps
		}
	}

	return usr, nil
}
