package b

import (
	"time"
)

// 1:N - Org : User
// M:N - User : Permission; additional "granted by" info

type Org struct {
	Key          string `json:"key"     orm:"pk;size(128)"`
	Name         string `json:"name"    orm:"size(64)"`
	ContactEmail string `json:"contact" orm:"size(128)"`
}

type User struct {
	Key         string        `json:"key"          orm:"pk;size(128)"`
	Login       string        `json:"login"        orm:"size(64)"`
	Email       string        `json:"email"        orm:"size(128)"`
	Deleted     bool          `json:"deleted"      orm:"default(false)"` // b/c unable to model DeletedTS as a pointer
	DeletedTS   time.Time     `json:"deleted_at"   orm:"null;type(datetime);default(null);column(deleted_ts)"`
	Org         *Org          `json:"org"          orm:"rel(fk)"`
	Permissions []*Permission `json:"permsissions" orm:"rel(m2m)"`
}

// Can specify uniqueness constraints (and also custom table names)
func (_ *User) TableUnique() [][]string {
	return [][]string{
		{"Key", "Login"},
		{"Email"},
	}
}

/*
	Attempted to use a custom many-to-many relationship table to track additional
	metadata but the pair of non-null fields seemed to break some assumptions
	within Beego if using the QueryM2Mer interface.

	Unless I missed something big I think you'd need to manually manage addition
	of associations which is kind af a pain.

	Was using this to join them:

	  rel_through(github.com/falun/hellorm/b.UserPermissions)

	And this association definition:

	  type UserPermissions struct {
	    Id         int64       `json:"id"          orm:"pk"`
	    Permission *Permission `json:"permission"  orm:"rel(fk)"`
	    User       *User       `json:"user"        orm:"rel(fk)"`
	    AssignedBy *User       `json:"assigned_by" orm:"rel(fk)"`
	  }

	  func (_ *UserPermissions) TableIndex() [][]string {
	    return [][]string{
	      {"AssignedBy"},
	    }
	  }
*/
type Permission struct {
	Id          int64  `json:"id"          orm: "pk"`
	Name        string `json:"name"        orm:"size(32)"`
	Description string `json:"description" orm:"size(256)"`
}
