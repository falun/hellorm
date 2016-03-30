package ryo

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
	Key          string `json:"key"`
	Name         string `json:"name"`
	ContactEmail string `json:"contact"`
}

type User struct {
	Key         string        `json:"key"`
	Login       string        `json:"login"`
	Email       string        `json:"email"`
	DeletedAt   *time.Time    `json:"deleted_at"`
	OrgKey      string        `json:"-"`
	Org         *Org          `json:"org"`
	Permissions *[]Permission `json:"permsissions"`
	Checksum    string        `json:"checksum"`
}

type Permission struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserPermission struct {
	UserKey      string
	PermissionId int64
	GrantedBy    string
}
