## <a name="nih"></a>NIH

Might as well roll my own on the off chance it's not too bad. Major benefits
being that I can keep my framework free of fucking panics.

## sqlx

Isn't really an ORM but some sugar for database/sql - [github](https://github.com/jmoiron/sqlx).
Bonus points in that jmoiron hangs out in #go-nuts and seems generally
friendly.

## <a name="gorm"></a>gorm

*In Progress*

The docs are somewhat disorienting. I believe it's because they seem to be a
collection of examples instead of explanation of principles. That could just
be the initial reading or the ones I focused on in the beginning.

I can't tell if the docs are just a prettier repackaging of godocs. I'm hoping
not becaue the comments there are ... underwhelming. I basically hoping that
with time and familiarity it becomes reasonable as I like it the most thus far.

### Error Handling

A bit weird but functional -- each DB operation returns a new `gorm.DB`. This
may be used to chain subsequent operations but also carries the results of your
previous operation:

    type DB struct {
      Value        interface{}
      Error        error
      RowsAffected int64
    }

It has the same mechanics of "panic on non-pointers" as everything else which
is disappointing.


### Model Definition / Migration
Does the *right thing* for most names.

Get `CreatedAt`, `UpdatedAt`, and `DeletedAt` for free if they're set on the
struct. If `DeletedAt` is present then soft deletion is the default behavior.

Similar to *beego* I don't think this generates FK constraints at the DDL level
and, instead, just uses that information to inform query construction. On the
other hand this supports adding the keys programatically. Which should be
sufficient if we need to generate DDL for one offs (like these experiments).
TODO: confirm the struct tags are sufficient for query planning.

Association table construction is confusing via struct tags. It looks like you
can work around this through the relatively tweakable representations that gorm
provides to it's relationship model (see [#653](https://github.com/jinzhu/gorm/issues/653)
for an example).

While realizing that model definition via framework isn't essential for DDL
management an accurate representation is important for query generation. This
means that the issues around the ability to accurately encode this into a single
point of our code has correctness and simplicity of our codebase.

## <a name="gorp"></a>gorp

*gorp* is a small library focused on mapping structs into a relational database.
It eschews trying to do anything with relationships. The version evaluated here
is v1 as determined by [gopkg.in/gorp.v1](http://gopkg.in/gorp.v1).

### Annotations
The first thing that strikes me when compared to Beego is that there is limited
support for struct tags in v1. That said it looks unlikely we'll be able to use
derived SQL to manage our tables so it's not clear this is much of a loss.

There is expanded support off master / v2 for update configurations via struct
tags but it seems like v2 is also a pretty significant rewrite and hasn't been
finalized so I'm hesitant to use it.

### Community / Project Velocity
A major concern is that there seems to be relatively little activity on gorp.
Feature development for v1 halted 2015-07 with a 3 phase process outilned in
[issue #270](https://github.com/go-gorp/gorp/issues/270) describing the plan
on getting to v2. It has seen some limited movement but still seems to be in
the first third with no planned exit.

### Optimistic Locking
In theory this is supported by the framework but it works only with int values
and panics when attempting to use a string. We could work aronud this by
changing our checksum into a version number, fall back to raw SQL (though this 
makes it kind of silly to wrap gorp around our code, or back port the change
in pre-v2 that supports non-int version fields.

### Pre/Post Operation Hooks 
Reasonable support for binding additional operations into insert/delete/etc
operations.

### Expressiveness
The exposed operations are *exceptionally* limited. It seems like you basically
get the core CRUD ops in with N cardinality before you're back to writing raw
SQL.

There is support for reading query results into a struct as long as your
projected attributes map cleanly to the struct. This is a thin layer of sugar
over "NewFromQueryResults" and seems prone runtime errors.

### Testing
Nothing out of the box; likely need to write a database dialect that mocks
things.

### Closing Commentary
*gorm* seems nice for what it does but the ideal audience feels like it's
people building relatively simple (or extremely complex) services that would
like some sugar around the basics but nothing between them in the actual SQL
in most cases.

I really like the fact that it's logging is `log.Logger` compatible as we can
just assign it to a `db` topic and call it done.

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
