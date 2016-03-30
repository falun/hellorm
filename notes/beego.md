## <a name="beego"></a>Beego

Beego is a framework targeting the whole suite of problems involved in building
a go-backed webapp. There is structure for defining models, views, routes,
etcetcyak. 

It lives [here](http://beego.me/).

While the whole basket of goodies is neat it's also too much for what we want
(and probably isn't worth switching to at this point anyway) but it seems to be
built in a well modularized manner so that bits can be pulled out and used in
isolation. I took a look at what its ORM provides when dealing with some roughly
realistic models.

The ORM specific docs are split up but the most important bits:

* [ORM basics](http://beego.me/docs/mvc/model/orm.md)
* [Model definition](http://beego.me/docs/mvc/model/models.md)
* [HOWTO CRUD](http://beego.me/docs/mvc/model/object.md)
* [Advanced Queries](http://beego.me/docs/mvc/model/query.md)
* [godocs](https://godoc.org/github.com/astaxie/beego/orm)

### Nil pointers

When attempting to have a nullable timestamp in a model the model registration
panicked. This means in order to have a deletion TS I can't just use existence
to determine if the user is gone and will need to pair it with a bool:

    Deleted   bool      `json:"deleted"    orm:"default(false)"`
    DeletedTS time.Time `json:"deleted_at" orm:"null;type(datetime);default(null);column(deleted_ts)"`


### Foreign Key constraints

It doesn't seem to know to encode FK constraints when generating DB schema.

    type Org struct {
      Key          string `json:"key"     orm:"pk;size(128)"`
      ...
    }
    
    type User struct {
      Key       string    `json:"key"        orm:"pk;size(128)"`
      ...
      Org       *Org      `json:"org"        orm:"rel(fk)"`
    }

The above generates a schema that doesn't include a FK constraint:

    CREATE TABLE IF NOT EXISTS "org" (
      "key" varchar(128) NOT NULL PRIMARY KEY,
      ...
    );
    
    CREATE TABLE IF NOT EXISTS "user" (
      "key" varchar(128) NOT NULL PRIMARY KEY,
      ...
      "org_id" varchar(128) NOT NULL,
    );

While we could manually add this (since we're not going ot use beego's db
migration abilities) but it seems not ideal.

Tracked as [issue 267](https://github.com/astaxie/beego/issues/267).

### Conditional updates

Possible.

    // build query condition
    cond := o.NewCondition().And("checksum", old_checksum)

    // establesh what we're querying
    o.QueryTable(&model.User{}).
      // attach the condition
      SetCond(cond).
      // or 'Filter("checksum", old_checksum)'
      // perform the update
      Update(o.Params{"name":new_name_value})

### Many To Many

Supported but weird / limited / opinionated.

If you take the route of providing a custom association table it's possible
to break the existing `QueryM2M` object.

    type UserPermissions struct {
      Id         int64       `json:"id"          orm:"pk"`
      Permission *Permission `json:"permission"  orm:"rel(fk)"`
      User       *User       `json:"user"        orm:"rel(fk)"`
      AssignedBy *User       `json:"assigned_by" orm:"rel(fk)"`
    }

Because there are two references to the user table in the above example and,
afaict, Beego doesn't provide ways to specify additional metadata it's not
possible (or at least not obvious how) to use the built in
`QueryM2M.Add(interface{})` approcah to creating entries.

### Transactions

Supported. You get a basic `Begin`, `Commit`, `Rollback` interface. Global per
`Ormer` object.

### Testing
Unclear. Probably test against a DB locally; maybe use SQLite for CI?

### Closing commentary

Some concerns include:

- *safety*     If you call (at least) `Ormer.Insert,Read` with a non-pointer it will panic. I guess this is no worse than the golang json interface?
- *safety*     No enforcement of FK constraints; for SQL construction on lookup and app-level validation only
- *safety*     The filter format `x__y__$test` can panic if you pass a bad query in
- *usability*  If you insert an item with no auto-icrement it returns an error `no LastInsertId available`
- *usability*  related objects vaguely annoying (probably no worse than any other ORM)
- *usability*  In theory it can produce SQL but can't be directly used (because it plays to a least common denominator of DB features) or natively manage its own migrations.
- *efficiency* curious query construction:  
	    o.QueryTable(&b.User{}).
	    	Filter("key", u.Key).
	    	Update(orm.Params{"email": "partial@example.com"})
I would expect to produce:

		 UPDATE "user"
		 SET "email" = 'partial@example.com'
		 WHERE "key" = 'd0c103c4-e471-48f7-5447-9e828bdc12b5'
but instead it produced:
    	
    	UPDATE "user"
    	SET "email" = 'partial@example.com'
    	WHERE "key" IN (
    		SELECT T0."key"
    		FROM "user" T0
    		WHERE T0."key" = 'd0c103c4-e471-48f7-5447-9e828bdc12b5'
    	)
