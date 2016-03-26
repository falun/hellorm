package main

import (
	"flag"
	"fmt"
	"github.com/falun/hellorm/b"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	"github.com/turbinelabs/tbn/idgen"
)

func init() {
}

var (
	mkID      = idgen.NewUUID()
	orgID     = "9cd24183-f848-48f8-6f55-0f07240700b9"
	falunID   = "b1d2f07b-5eed-45e8-4bdb-bcbcdec4e05e"
	altID     = "d0c103c4-e471-48f7-5447-9e828bdc12b5"
	readPerm  = b.Permission{Id: 1}
	writePerm = b.Permission{Id: 2}
	delPerm   = b.Permission{Id: 3}
)

func insertNewItem() {
	newKey, _ := mkID()
	key := string(newKey)

	org := &b.Org{Key: key, Name: "SomeCo", ContactEmail: "admin@someco.io"}

	orm := orm.NewOrm()
	n, err := orm.Insert(org)
	fmt.Printf("n: %+v\n", n)
	fmt.Printf("err: %T: %+v\n", err, err)
}

// full updates are actually pretty trivial
func fullUpdateObject() {
	orm := orm.NewOrm()

	u := b.User{Key: altID}
	err := orm.Read(&u)

	// change e-mail
	u.Email = "newemail@example.com"
	orm.Update(&u)

	err = orm.Read(&u)
	fmt.Printf("user: %+v\n", u)
	fmt.Printf("err: %T: %+v\n", err, err)
}

func conditionalAndPartialUpdateObject() {
	o := orm.NewOrm()

	u := b.User{Key: altID}

	p := orm.Params{"email": "partial@example.com"}

	// style one -- use filter to construct inline which will get ANDed together
	r, err := o.
		QueryTable(&b.User{}).
		Filter("key", u.Key).                      // key must match
		Filter("email__iendswith", "example.com"). // email field must end with
		Update(p)

	fmt.Printf("r: %+v\n", r)
	fmt.Printf("err: %T: %+v\n", err, err)

	p["email"] = "test2@example.com"

	// style two -- construct explicit conditions; this allows you to do more
	// complex clauses (n.b. this is a nonsensical query, don't try to understand
	// why it does what it does)
	c := orm.NewCondition().
		And("key", u.Key).
		And("email__iendswith", "example.com").
		// and the filter format lets you descend into related objects to add
		// tests; this will do a join against permisions for users who have a
		// permission in a set
		And("permissions__permission_id__in", delPerm.Id, readPerm.Id)
	r, err = o.QueryTable(&b.User{}).SetCond(c).Update(p)

	fmt.Printf("r: %+v\n", r)
	fmt.Printf("err: %T: %+v\n", err, err)
}

func queryAndLoadRelatedObjects() {
	printUser := func(u *b.User, err error) {
		fmt.Printf("err: %T: %+v\n", err, err)
		fmt.Printf("user: %+v\n", u)
		fmt.Printf("org: %+v\n", u.Org)
		fmt.Println("permissions: {")
		for _, v := range u.Permissions {
			fmt.Println("  -", v)
		}
		fmt.Println("}")
		fmt.Println()
		fmt.Println()
	}

	o := orm.NewOrm()
	u := b.User{Key: altID}

	// simple read doesn't get it:
	err := o.Read(&u)
	printUser(&u, err)

	// args for the LoadRelated call are fucked up:
	// 1 - model to load
	// 2 - name of the field to load
	// 3 - MAGIC:
	//     [0] bool   true useDefaultRelsDepth ; false  depth 0
	//     [0] int    loadRelationDepth
	//     [1] int    limit default limit 1000
	//     [2] int    offset default offset 0
	//     [3] string order  for example : "-Id"
	n, err := o.LoadRelated(&u, "Org", 1)
	fmt.Printf("n: %T: %+v\n", n, n)
	printUser(&u, err)

	n, err = o.LoadRelated(&u, "Permissions", 1)
	fmt.Printf("n: %T: %+v\n", n, n)
	printUser(&u, err)
}

func main() {
	var (
		dbUser = ""
		dbPass = ""
		dbHost = ""
		db     = ""
	)

	flag.StringVar(&dbUser, "u", "", "set database username")
	flag.StringVar(&dbPass, "p", "", "set database password")
	flag.StringVar(&dbHost, "h", "localhost", "host running database")
	flag.StringVar(&db, "d", "", "set database")

	orm.RegisterDriver("postgres", orm.DRPostgres)
	orm.RegisterDataBase(
		"default",
		"postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, db))

	orm.Debug = true
	orm.DefaultTimeLoc = time.UTC
	orm.RegisterModel(
		new(b.Org),
		new(b.User),
		new(b.Permission),
	)

	fmt.Println("-")
	orm.RunCommand()
}
