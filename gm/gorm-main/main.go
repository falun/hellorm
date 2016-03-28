package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/falun/hellorm"
	"github.com/falun/hellorm/gm"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
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

// configure gorm connection to DB
func initDB() *gorm.DB {
	dburl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, db)
	db, err := gorm.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}

	db.LogMode(true)

	return db
}

func initModels(db *gorm.DB) {
}

func dropTables(db *gorm.DB) {
	// r := db.DropTableIfExists(&gm.UserPermission{})
	r := db.DropTableIfExists("user_permissions")
	if r.Error != nil {
		fmt.Println(r.Error)
	}

	r = db.DropTableIfExists(&gm.Permission{})
	if r.Error != nil {
		fmt.Println(r.Error)
	}

	r = db.DropTableIfExists(&gm.User{})
	if r.Error != nil {
		fmt.Println(r.Error)
	}

	r = db.DropTableIfExists(&gm.Org{})
	if r.Error != nil {
		fmt.Println(r.Error)
	}
}

// attach our models to gorm
func createTables(db *gorm.DB) {
	db.CreateTable(
		&gm.Org{},
		//&gm.UserPermission{},
		&gm.User{},
		&gm.Permission{},
	)

	u := &gm.User{}
	db.Model(u).AddForeignKey("org_key", "orgs(key)", "RESTRICT", "CASCADE")

	/*
		up := &gm.UserPermission{}
		db.Model(up).AddForeignKey("permission_id", "permissions(id)", "RESTRICT", "RESTRICT")
		db.Model(up).AddForeignKey("user_key", "users(key)", "RESTRICT", "CASCADE")
	*/
}

var (
	mkID      = hellorm.MkID
	orgID     = "9cd24183-f848-48f8-6f55-0f07240700b9"
	falunID   = "b1d2f07b-5eed-45e8-4bdb-bcbcdec4e05e"
	altID     = "d0c103c4-e471-48f7-5447-9e828bdc12b5"
	readPerm  = gm.Permission{Id: 1, Name: "read", Description: "can read data"}
	writePerm = gm.Permission{Id: 2, Name: "write", Description: "can write data"}
	delPerm   = gm.Permission{Id: 3, Name: "del", Description: "can delete data"}
)

func setupCoreItems(db *gorm.DB) {
	org := &gm.Org{Key: orgID, Name: "SomeCo", ContactEmail: "heythere@someco.io"}
	user := &gm.User{Key: falunID, Login: "falun", Email: "test@asnoteh", OrgKey: orgID, Checksum: mkID()}
	user2 := &gm.User{Key: altID, Login: "alt", Email: "altemail@asnoteh", OrgKey: orgID, Checksum: mkID()}

	/*
		up1 := &gm.UserPermission{altID, readPerm.Id, falunID}
		up2 := &gm.UserPermission{altID, writePerm.Id, falunID}
	*/
	user2.Permissions = []gm.Permission{readPerm, writePerm}

	items := []interface{}{org, user, user2, &readPerm, &writePerm, &delPerm /*, up1, up2*/}
	for _, i := range items {
		result := db.Create(i)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
	}
}

func insertNewItem(db *gorm.DB) {
	org := gm.Org{Key: mkID(), Name: "Fancy Company", ContactEmail: "admin@fancy.net"}
	r := db.Create(&org)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	v := r.Value.(*gm.Org)
	fmt.Printf("%T - key: %s\n", v, v.Key)
}

func fullUpdateObject(db *gorm.DB) {
	// read the existing data for the trivial/full update
	u := &gm.User{Key: altID}
	r := db.Find(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	// update
	u.Login = "alt-login"

	// and write
	r = db.Save(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	u = r.Value.(*gm.User)
	fmt.Printf("%T : %+v\n", u, *u)
}

func queryAndLoadRelatedObjects(db *gorm.DB) {
	u := &gm.User{Key: altID}

	// load related fields
	r := db.Preload("Org").Preload("Permissions").Find(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	fmt.Printf("user: %+v\n", u)
	fmt.Printf("org: %+v\n", u.Org)
	fmt.Printf("permissions: %+v\n", u.Permissions)

	u.Permissions = []gm.Permission{readPerm, delPerm}
	r = db.Save(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	// for giggles how do we delete a related field?
	ra := db.Model(u).Association("Permissions").Delete(readPerm, writePerm)
	if ra.Error != nil {
		log.Fatal(ra.Error)
	}

	fmt.Printf("user: %+v\n", u)
	fmt.Printf("org: %+v\n", u.Org)
	fmt.Printf("permissions: %+v\n", u.Permissions)
}

func partialUpdateObject(db *gorm.DB) {
	u := &gm.User{Key: altID}

	// load related fields
	r := db.Preload("Org").Preload("Permissions").Find(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	fmt.Printf("user: %+v\n", u)
	fmt.Printf("org: %+v\n", u.Org)
	fmt.Printf("permissions: %+v\n", u.Permissions)

	db.Model(u).Select("login").Updates(map[string]interface{}{
		"login": "alt-login",
	})
}

func conditionalQuery(db *gorm.DB) {
	u := &gm.User{Key: altID}

	// load related fields
	r := db.Preload("Org").Preload("Permissions").Find(u)
	if r.Error != nil {
		log.Fatal(r.Error)
	}

	tgtCS := u.Checksum + "A"

	r = db.Model(u).
		Select("login").
		Where(map[string]interface{}{"checksum": tgtCS}).
		Updates(map[string]interface{}{"login": "conditional-login"})

	if r.Error != nil {
		log.Fatal(r.Error)
	}

	fmt.Printf("Updated %d rows\n", r.RowsAffected)
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
	initModels(db)

	if dropFirst {
		dropTables(db)
	}
	if create {
		createTables(db)
	}
	if core {
		setupCoreItems(db)
	}

	fmt.Println()
	fmt.Println()

	switch ex {
	case "insert":
		insertNewItem(db)
	case "update":
		fullUpdateObject(db)
	case "partial-update":
		partialUpdateObject(db)
	case "conditional":
		conditionalQuery(db)
	case "related":
		queryAndLoadRelatedObjects(db)
	default:
		fmt.Println("No example executed")
	}
}
