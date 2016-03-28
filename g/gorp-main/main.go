package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/falun/hellorm"
	"github.com/falun/hellorm/g"

	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

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

// configure gorp connection to DB
func initDB() *gorp.DbMap {
	dburl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, db)
	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "", log.Lmicroseconds))

	return dbmap
}

// attach our models to gorp
func initTables(gdb *gorp.DbMap) {
	orgs := gdb.AddTableWithName(g.Org{}, "orgs")
	orgs.ColMap("Key").SetMaxSize(128).SetNotNull(true).SetUnique(true)
	orgs.ColMap("Name").SetMaxSize(64).SetNotNull(true)
	orgs.ColMap("ContactEmail").SetMaxSize(64).SetNotNull(false)
	orgs.SetKeys(false, "key")

	users := gdb.AddTableWithName(g.User{}, "users")
	users.SetKeys(false, "key")
	users.ColMap("Login").SetUnique(true)

	permissions := gdb.AddTableWithName(g.Permission{}, "permissions")
	permissions.SetKeys(true, "id")

	userPermAssoc := gdb.AddTableWithName(g.UserPermission{}, "user_permissions")
	userPermAssoc.SetUniqueTogether("user_key", "permission_id")
}

var (
	mkID      = hellorm.MkID
	orgID     = "9cd24183-f848-48f8-6f55-0f07240700b9"
	falunID   = "b1d2f07b-5eed-45e8-4bdb-bcbcdec4e05e"
	altID     = "d0c103c4-e471-48f7-5447-9e828bdc12b5"
	readPerm  = g.Permission{Id: 1, Name: "read", Description: "can read data"}
	writePerm = g.Permission{Id: 2, Name: "write", Description: "can write data"}
	delPerm   = g.Permission{Id: 3, Name: "del", Description: "can delete data"}
)

func setupCoreItems(db *gorp.DbMap) {
	org := &g.Org{Key: orgID, Name: "SomeCo", ContactEmail: "heythere@someco.io"}
	user := &g.User{Key: falunID, Login: "falun", Email: "test@asnoteh", OrgKey: orgID, Checksum: mkID()}
	user2 := &g.User{Key: altID, Login: "alt", Email: "altemail@asnoteh", OrgKey: orgID, Checksum: mkID()}
	up1 := &g.UserPermission{altID, readPerm.Id, falunID}
	up2 := &g.UserPermission{altID, writePerm.Id, falunID}

	err := db.Insert(org, user, user2, &readPerm, &writePerm, &delPerm, up1, up2)
	if err != nil {
		log.Fatal(err)
	}
}

func insertNewItem(db *gorp.DbMap) {
	org := g.Org{Key: mkID(), Name: "Fancy Company", ContactEmail: "admin@fancy.net"}
	err := db.Insert(&org)
	if err != nil {
		log.Fatal(err)
	}
}

func fullUpdateObject(db *gorp.DbMap) {
	u, err := db.Get(g.User{}, altID)
	if err != nil {
		log.Fatal(err)
	}

	cast := u.(*g.User)
	cast.Email = "newmail@mail.mail"

	n, err := db.Update(cast)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(n, "updates made")
}

func queryAndLoadRelatedObjects(db *gorp.DbMap) {
	u, err := db.Get(g.User{}, altID)
	if err != nil {
		log.Fatal(err)
	}

	cast := u.(*g.User)

	// because gorm doesn't support relations at the top level we need to
	// drop down into SQL to load related fields efficently. In order to keep
	// from having to do multiple queries (1: user -> perm ids; 2: perm ids ->
	// data) we have to project field names into how they're exposed in the go
	// struct (or a similar solution).
	query := fmt.Sprintf(`
		SELECT p.id "Id" , p.name "Name", p.description "Description"
		FROM
			permissions p INNER JOIN user_permissions up
			ON p.id = up.permission_id
		WHERE
			up.user_key = '%s'
	`, altID)

	r, err := db.Select(&g.Permission{}, query)
	if err != nil {
		log.Fatal(err)
	}

	// and then we need to cast them into the appropriate type
	rc := make([]g.Permission, 0, len(r))
	for _, e := range r {
		if e == nil {
			continue
		}
		rc = append(rc, *e.(*g.Permission))
	}

	// and finally update.
	cast.Permissions = rc
	fmt.Printf("%+v\n", cast)

	// It works but, on the whole, this is pretty miserable.
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
	dbMap := initDB()
	initTables(dbMap)
	if dropFirst {
		dbMap.DropTables()
	}
	if create {
		err := dbMap.CreateTables()
		if err != nil {
			log.Fatal(err)
		}
	}
	if core {
		setupCoreItems(dbMap)
	}

	fmt.Println()
	fmt.Println()

	switch ex {
	case "insert":
		insertNewItem(dbMap)
	case "update":
		fullUpdateObject(dbMap)
	case "related":
		queryAndLoadRelatedObjects(dbMap)
	default:
		fmt.Println("No example executed")
	}
}
