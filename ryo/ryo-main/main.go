package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/falun/hellorm"
	"github.com/falun/hellorm/ryo"
	"github.com/falun/hellorm/ryo/orgsql"
	"github.com/falun/hellorm/ryo/permsql"
	"github.com/falun/hellorm/ryo/userpermsql"
	"github.com/falun/hellorm/ryo/usersql"

	_ "github.com/lib/pq"
)

type ryorm struct {
	Db       *ryo.DB
	Org      *orgsql.Org
	User     *usersql.User
	Perm     *permsql.Permission
	UserPerm *userpermsql.UserPermission
}

var (
	dbUser = ""
	dbPass = ""
	dbHost = ""
	db     = ""
)

func init() {
	flag.StringVar(&dbUser, "u", "", "set database username")
	flag.StringVar(&dbPass, "p", "", "set database password")
	flag.StringVar(&dbHost, "h", "localhost", "host running database")
	flag.StringVar(&db, "d", "", "set database")
}

// configure connection to DB
func initDB() *sql.DB {
	dburl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, db)
	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func initModels(orm *ryorm) {
}

func dropTables(orm *ryorm) {
	do := func(q string) {
		_, e := orm.Db.Exec(q)
		if e != nil {
			log.Println(e)
		}
	}

	do(userpermsql.DropTable)
	do(permsql.DropTable)
	do(usersql.DropTable)
	do(orgsql.DropTable)
}

// attach our models to sql
func createTables(orm *ryorm) {
	do := func(q string) {
		_, e := orm.Db.Exec(q)
		if e != nil {
			log.Println(e)
		}
	}

	do(orgsql.CreateTable)
	do(usersql.CreateTable)
	do(permsql.CreateTable)
	do(userpermsql.CreateTable)
}

var (
	mkID      = hellorm.MkID
	orgID     = "9cd24183-f848-48f8-6f55-0f07240700b9"
	falunID   = "b1d2f07b-5eed-45e8-4bdb-bcbcdec4e05e"
	altID     = "d0c103c4-e471-48f7-5447-9e828bdc12b5"
	readPerm  = ryo.Permission{Name: "read", Description: "can read data"}
	writePerm = ryo.Permission{Name: "write", Description: "can write data"}
	delPerm   = ryo.Permission{Name: "del", Description: "can delete data"}
)

func setupCoreItems(orm *ryorm) {
	org := ryo.Org{Key: orgID, Name: "SomeCo", ContactEmail: "heythere@someco.io"}
	user := ryo.User{Key: falunID, Login: "falun", Email: "test@asnoteh", OrgKey: orgID, Checksum: mkID()}
	user2 := ryo.User{Key: altID, Login: "alt", Email: "altemail@asnoteh", OrgKey: orgID, Checksum: mkID()}

	do := func(action func() error) {
		e := action()
		if e != nil {
			log.Fatal(e)
		}
	}

	do(func() error { return orm.Org.Insert(org) })
	do(func() error { return orm.User.Insert(user) })
	do(func() error { return orm.User.Insert(user2) })

	doPerm := func(perm *ryo.Permission) {
		p, e := orm.Perm.Insert(*perm)
		if e != nil {
			log.Fatal(e)
		}
		perm.Id = p.Id
	}

	doPerm(&readPerm)
	doPerm(&writePerm)
	doPerm(&delPerm)

	orm.UserPerm.Grant(readPerm, user2, user)
	orm.UserPerm.Grant(writePerm, user2, user)
}

func insertNewItem(orm *ryorm) {
	org := ryo.Org{Key: mkID(), Name: "Fancy Company", ContactEmail: "admin@fancy.net"}
	e := orm.Org.Insert(org)
	if e != nil {
		log.Fatal(e)
	}
}

func fullUpdateObject(db *ryorm) {
	u, e := db.User.Get(altID, nil)
	if e != nil {
		log.Fatal(e)
	}

	oldCS := u.Checksum
	u.Email = "newmail@mail.mail"
	u.Checksum = mkID()

	n, e := db.User.Update(u, oldCS)
	if e != nil {
		log.Fatal(e)
	}
	log.Println(n, "rows affected")
}

func queryAndLoadRelatedObjects(db *ryorm) {
	u, e := db.User.Get(altID, &usersql.GetOpts{LoadOrg: true, LoadPermissions: true})
	if e != nil {
		log.Fatal(e)
	}

	b, e := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(b))
}

func partialUpdateObject(db *ryorm) {
	u, e := db.User.Get(altID, nil)
	if e != nil {
		log.Fatal(e)
	}

	u.Login = "alt-login"
	n, e := db.User.PartialUpdate(u, nil, usersql.ColLogin, usersql.ColEmail)
	if e != nil {
		log.Fatal(e)
	}
	log.Println(n, "rows affected")
}

func conditionalQuery(db *ryorm) {
	u, e := db.User.Get(altID, nil)
	if e != nil {
		log.Fatal(e)
	}

	oldCS := u.Checksum
	u.Login = "alt-alt"
	u.Checksum = mkID()
	n, e := db.User.PartialUpdate(
		u, &usersql.UpdateOpts{true, oldCS}, usersql.ColChecksum, usersql.ColLogin)
	if e != nil {
		log.Fatal(e)
	}
	log.Println(n, "rows affected")
}

func main() {
	dropFirst := false
	flag.BoolVar(&dropFirst, "drop", false, "drop tables first")

	create := false
	flag.BoolVar(&create, "create", false, "create tables")

	core := false
	flag.BoolVar(&core, "core", false, "insert core shit")

	ex := ""
	flag.StringVar(&ex, "ex", "", "which example to run")

	flag.Parse()
	db := initDB()

	// lolwiring
	var (
		rdb  = &ryo.DB{db}
		up   = &userpermsql.UserPermission{rdb}
		perm = &permsql.Permission{rdb}
		org  = &orgsql.Org{rdb}
		orm  = &ryorm{
			Db:       rdb,
			Org:      org,
			User:     &usersql.User{rdb, org, perm, up},
			Perm:     perm,
			UserPerm: up,
		}
	)

	initModels(orm)

	if dropFirst {
		dropTables(orm)
	}
	if create {
		createTables(orm)
	}
	if core {
		setupCoreItems(orm)
	}

	fmt.Println()
	fmt.Println()

	switch ex {
	case "insert":
		insertNewItem(orm)
	case "update":
		fullUpdateObject(orm)
	case "partial-update":
		partialUpdateObject(orm)
	case "conditional":
		conditionalQuery(orm)
	case "related":
		queryAndLoadRelatedObjects(orm)
	default:
		fmt.Println("No example executed")
	}
}
