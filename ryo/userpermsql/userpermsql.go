package userpermsql

import (
	"fmt"

	"github.com/falun/hellorm/ryo"
)

var (
	TableName = "user_permissions"

	DropTable   = fmt.Sprintf(`DROP TABLE %q`, TableName)
	CreateTable = fmt.Sprintf(`
    CREATE TABLE %q (
      permission_id bigint NOT NULL,
      user_key character varying(128) NOT NULL,
      assigned_by character varying(128) NOT NULL,

    CONSTRAINT user_permissions_pkey PRIMARY KEY (user_key, permission_id),
    CONSTRAINT user_permissions_user_key_user_key_foreign_key FOREIGN KEY (user_key)
      REFERENCES users(key) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT user_permissions_permission_id_permission_id_foreign FOREIGN KEY (permission_id)
      REFERENCES permissions(id) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT user_permissions_assigned_by_user_key FOREIGN KEY (assigned_by)
      REFERENCES users(key) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE RESTRICT
  )`, TableName)

	Insert = fmt.Sprintf(
		`INSERT INTO %q ("permission_id", "user_key", "assigned_by") VALUES($1, $2, $3)`,
		TableName)
)

func insert(db *ryo.DB, permId int64, assignedTo, assignedBy string) error {
	_, e := db.Exec(Insert, permId, assignedTo, assignedBy)
	return e
}

type UserPermission struct{ Db *ryo.DB }

func (u *UserPermission) Grant(perm ryo.Permission, toUser, byUser ryo.User) error {
	return insert(u.Db, perm.Id, toUser.Key, byUser.Key)
}
