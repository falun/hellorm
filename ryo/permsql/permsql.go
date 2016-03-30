package permsql

import (
	"fmt"

	"github.com/falun/hellorm/ryo"
	"github.com/falun/hellorm/ryo/userpermsql"
)

var (
	TableName = "permissions"

	DropTable   = fmt.Sprintf(`DROP TABLE %q`, TableName)
	CreateTable = fmt.Sprintf(`
    CREATE TABLE %q (
      id bigserial NOT NULL,
      name text,
      description text,
      CONSTRAINT permissions_pkey PRIMARY KEY (id)
    )`, TableName)

	ReadByKey = fmt.Sprintf(`SELECT "id", "name", "description" FROM %q WHERE id = $1`, TableName)
	Insert    = fmt.Sprintf(`
	  INSERT INTO %q ("name", "description") VALUES($1, $2) RETURNING %q."id"
	`, TableName, TableName)
	// May not want to support this one as it doesn't impact the sequence generator
	InsertWithId = fmt.Sprintf(`INSERT INTO %q ("id", "name", "description") VALUES($1, $2, $3)`, TableName)

	GrantedToUser = fmt.Sprintf(`
    SELECT p."id", p."name", p."description"
    FROM %q AS p INNER JOIN %q AS up
      ON p."id" = up."permission_id"
    WHERE up."user_key" = $1`, TableName, userpermsql.TableName)
)

func insert(db *ryo.DB, perm ryo.Permission) (ryo.Permission, error) {
	if perm.Id != 0 {
		_, e := db.Exec(InsertWithId, perm.Id, perm.Name, perm.Description)
		return perm, e
	}

	var id int64
	e := db.QueryRow(Insert, perm.Name, perm.Description).Scan(&id)
	if e != nil {
		return perm, e
	}

	perm.Id = id
	return perm, nil
}

type Permission struct{ Db *ryo.DB }

func (u *Permission) Insert(newPerm ryo.Permission) (ryo.Permission, error) {
	return insert(u.Db, newPerm)
}

func grantedTo(db *ryo.DB, userKey string) ([]ryo.Permission, error) {
	rows, err := db.Query(GrantedToUser, userKey)
	if err != nil {
		return nil, err
	}

	result := make([]ryo.Permission, 0)
	for rows.Next() {
		p := ryo.Permission{}
		err = rows.Scan(&p.Id, &p.Name, &p.Description)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

func (u *Permission) GrantedTo(user string) ([]ryo.Permission, error) {
	return grantedTo(u.Db, user)
}
