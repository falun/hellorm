package g

import (
	"time"
)

// Attempting to model the following objects:
//   Org - an organization
//   User - people within an org
//   Permissions - things people can do
//
// They will have relationships in this cardinality:
//   1:N - Org : User
//   M:N - User : Permission
// Additionally each permission bit should be trackable to who granted that bit
//
// And finally a User object should never be deleted; instead it carries
// a "DeletedTS" attribute which we can use to determine if they have an active
// account.

type Org struct {
	Key          string `json:"key"     db:"key"`
	Name         string `json:"name"    db:"name"`
	ContactEmail string `json:"contact" db:"contact_email"`
}

type User struct {
	Key         string       `json:"key"          db:"key"`
	Login       string       `json:"login"        db:"login"`
	Email       string       `json:"email"        db:"email"`
	DeletedTS   *time.Time   `json:"deleted_at"   db:"deleted_ts"`
	OrgKey      string       `json:"org"          db:"org_key"`
	Org         *Org         `json:"-"            db:"-"`
	Permissions []Permission `json:"permsissions" db:"-"`
	Checksum    string       `json:"checksum"     db:"checksum"`
}

type Permission struct {
	Id          int64  `json:"id"          db:"id"`
	Name        string `json:"name"        db:"name"`
	Description string `json:"description" db:"description"`
}

type UserPermission struct {
	UserKey   string `db:"user_key"`
	PermId    int64  `db:"permission_id"`
	GrantedBy string `db:"granted_by"`
}
