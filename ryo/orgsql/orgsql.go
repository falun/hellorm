package orgsql

import (
	"fmt"

	"github.com/falun/hellorm/ryo"
)

var (
	TableName = "orgs"

	DropTable   = fmt.Sprintf(`DROP TABLE %q`, TableName)
	CreateTable = fmt.Sprintf(`
    CREATE TABLE %q (
      key text NOT NULL,
      name text,
      contact_email text,
      CONSTRAINT orgs_pkey PRIMARY KEY (key)
    )`, TableName)

	ReadByKey = fmt.Sprintf(`SELECT "key", "name", "contact_email" FROM %q WHERE key = $1`, TableName)
	Insert    = fmt.Sprintf(`INSERT INTO %q ("key", "name", "contact_email") VALUES($1, $2, $3)`, TableName)
)

func insert(db *ryo.DB, org ryo.Org) error {
	_, e := db.Exec(Insert, org.Key, org.Name, org.ContactEmail)
	return e
}

type Org struct{ Db *ryo.DB }

func (o *Org) Insert(newOrg ryo.Org) error {
	return insert(o.Db, newOrg)
}

func (u *Org) Get(orgKey string) (ryo.Org, error) {
	org := ryo.Org{}
	err := u.Db.QueryRow(ReadByKey, orgKey).Scan(&org.Key, &org.Name, &org.ContactEmail)

	if err != nil {
		return ryo.Org{}, err
	}

	return org, nil
}
