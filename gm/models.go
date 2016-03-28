package gm

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
	Key          string `json:"key"     gorm:"primary_key"`
	Name         string `json:"name"    gorm:""`
	ContactEmail string `json:"contact" gorm:""`
}

/*
	Custom association table for permissions foiled again!

	In theory there is explicit support for this (based on docs
	http://jinzhu.me/gorm/associations.html#many-to-many ) but I was never able
	to make it work. Every time I used Preload it would fail to generate valid
	SQL. I'm 99% sure you can configure it in code but it surpassed the amount
	of code-reading I was willing for a "how does this work" survey.

	*shrug*

	type UserPermission struct {
	  UserKey      string `gorm:"primary_key;ForeignKey"`
	  PermissionId int64  `gorm:"primary_key"`
	  GrantedBy    string `gorm:""`
	}
*/
type User struct {
	Key         string       `json:"key"          gorm:"primary_key"`
	Login       string       `json:"login"        gorm:""`
	Email       string       `json:"email"        gorm:"" sql:"unique"`
	DeletedAT   *time.Time   `json:"deleted_at"   gorm:""`
	OrgKey      string       `json:"org"          gorm:""`
	Org         *Org         `json:"-"            gorm:"ForeignKey:OrgKey"`
	Permissions []Permission `json:"permsissions" gorm:"many2many:user_permissions"`
	Checksum    string       `json:"checksum"     gorm:""`
}

type Permission struct {
	Id          int64  `json:"id"          gorm:"primary_key"`
	Name        string `json:"name"        gorm:""`
	Description string `json:"description" gorm:""`
}
