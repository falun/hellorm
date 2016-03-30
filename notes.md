## <a name="ryo"></a>ryo
:sob:. Whelp, might as well roll my own while I'm doing this on the off chance
it's not too bad. Major benefits being that I can

* keep my library free of panics on bad input
* minimize spread of magic strings
* directly control SQL being used for various operations
* built for testability

*A day passes*

Okay, that was terrible. Well, specifically, the act of rebuilding the tooling
necessary to do this was terrible. Once the effort was done using my own infra
turned out to be at least as easy as utilizing existing ORMish libraries. Part
of this can be attributed just to familiarity and part to the fact that it was
built to support only the usecases I care about. My design / API was not super
thoughtful so I'm guessing there is a good bit of improvement that can be made
as well.

Package & struct names, var placement, and if I actually want/need to wrap
`sql.DB` and `sql.Tx` all need some actual thought.

### Model Definition / Schema Migration

None. No support. Nil. Zip. Zilch. Write your own damned DDL and use
[goose](https://bitbucket.org/liamstask/goose) or similar.

### Error Handling / Safety etc.

Only as good as my knowledge of the underlying `database/sql` (or equivalent)
library and laziness about wiring errors back. On the other hand 100% fewer
runtime panics at runtime. Additionally there is a much lower frequency of
getting `interface{}` in your call chain (used for object attribute extraction
which should be extremely testable). Everybody loves types and now your
database interactions can too!

### Complexity
Worth calling out: wiring the data access interfaces together is a *huge pain*.
Because each component may need references to one or more which may, in turn,
need references to the thing it's being embedded in this is kind of asking for
some kind of injection solution.  Instead I have:

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

Other than that caveat complexity level is pretty middle of the road. The code
itself is all super straight forward but in a real system there is going to be
a ton of it as using typed interfaces means you can't as easily generalize.

As mentioned above a *user* has relatively low complexity given the targeted
nature of the API and the compiler is able to catch most syntactic problems.
Logical issues (e.g., forgetting to update a given field): still your problem.

One major concern is that this approach makes it less simple to update
attributes that are stored separately from an object. The primary example of
this in my simplified model is the `User:Permission` relationship.  This could
obviously be handled by building that logic into parent object but it seems
like a lack of a general solution could yield an unfortunate quantity of code
that needs to be written. (Also it's not clear to me you always care about
updating child objects.)

It's also not possible in my implementation to join multiple sets of queries
together in a single transaction. This would obviously need to change; possibly
meaning an introduction of a new abstraction layer (something about a
`QueryExecutor`?) that `*sql.DB` and `*sql.Tx` meet (or whatever layer I wrap
around them).

One bit of noise that will be removed if this goes to prod is the use of
`ryo` for table management. This was included in this project for a more
complete comparison to the other frameworks (and because it made repated
testing *way* easier).

### Testability
Extermely high!

The `ryo` code is broken into two or three components (depending on how you
count):

1. SQL&mdash;at the lowest level we provide SQL for specific operations on
a domain object.
2. Operations&mdash;Above SQL is a layer that exposes logical operations
without specifying any particular implementation. This is the `usersql.User`
object and could easily be replaced with an interface to facilitate mulptile
implementations (`usersql.sqlUser` & `usersql.mockUser` for example).
3. Logic&mdash;Business logic that consumes the operations layer.

Each layer can be thourghly tested both locally and in CI without reliance on
the piece above or below it existing. (A bit of a footnote that the SQL layer
may be annoying in CI / if parallelizing your tests if you want a real DB
backing it but it's still doable.)

### Future Proofing

Thinking about potential for unexpected development requirements with bespoke
API and the expanded dev-time they demand is somewhat concerning. Part of the
benefit with the ORMish solutions is that out of the box they provide the
knowledge there is a way to do whatever you want... even if it's ugly.

One bright spot is that there are many points where it would be possible to
simplify maintenance through codegen since so much of this should look the same
varying only by SQL.

Migrating between multiple backend implementations should be pretty
uncomplicated as you can build it into the business logic layer or the data
access layer using standard a/b switching mechanics which retains testability
in the face of an evolving datastore.

## sqlx

Isn't really an ORM but some sugar for `database/sql` - [github](https://github.com/jmoiron/sqlx).
And nifty helpers to scan into structs. Bonus points in that jmoiron hangs out
in `#go-nuts` and seems generally helpful.

I expect I'll fold this into ryo if it goes to prod if I can figure out logging
as I have no desire to rewrap everything, again.

## <a name="gorm"></a>gorm
The docs are somewhat disorienting. I believe it's because they seem to be a
collection of examples instead of explanation of principles. That could just
be the initial reading or the ones I focused on in the beginning.

I can't tell if the docs are just a prettier repackaging of godocs. I'm hoping
not becaue the comments there are ... underwhelming. I'm basically holding out
for time and familiarity to make this reasonable as I like it the most thus
far.

### Error Handling

A bit weird but functional -- most operations return a new instance of that
type (`gorm.DB`, `gorm.Association`, etc.). This may be used to chain
subsequent operations but also carries the results of your previous operation:

    type DB struct {
      Value        interface{}
      Error        error
      RowsAffected int64
    }

It has the same mechanics of "panic on non-pointers" as everything else which
is disappointing.

### Model Definition / Migration

Does the "right thing" for most fields but doesn't seem to support tag-based
name specification.

Get `CreatedAt`, `UpdatedAt`, and `DeletedAt` for free if they're set on the
struct. If `DeletedAt` is present then soft deletion is the default behavior.

Similar to beego I don't think this generates FK constraints at the DDL level
and, instead, just uses that information to inform query construction. On the
other hand this supports adding the keys programatically. Which should be
sufficient if we need to generate DDL for one offs (like these experiments).

Association table construction is confusing via struct tags. It looks like you
can work around this through the relatively tweakable representations that gorm
provides to it's relationship model (see [#653](https://github.com/jinzhu/gorm/issues/653)
for an example).

While realizing that model definition via framework isn't essential for DDL
management an accurate representation is important for query generation. This
means that the issues around the ability to accurately encode this into a single
point of our code has correctness and simplicity of our codebase.

### Complex relationships
I touch on this above but it's worth calling out here: my initial toying around
has not yielded a fully functional complex relationship or one that I am super
happy with.  Spectifically:

* Custom association tables are not working
* Examples provided for annotation-based configuration seem lacking/incorrect.
As a result our relationship configuration is likely to be a weird mix of code
and tags (or just code). This is ... not ideal as the code-centric solution is
not the most clear API in the world (from what I can tell).
* Updating with the default `Save` behavior is append only for m2m relations
and issues an unfortunate number of writes. A Simple "add permission" became
4 updates (users, 2 permissions, org) and 2 inserts (2 user\_permissions) with
nested select (for the `WHERE NOT EXISTS` protection). This grows linearly as
there is a write for each of the associated permission objects. This pretty
much forces you into selecting fields to be written manually or turning off
the `gorm:write_associations` flag. The former is just annoying and the latter
creates a slightly weird disconnect with how top level and nested objects get
updated.

### Final thoughts
> WARNING When delete a record, you need to ensure it's primary field has
> value, and GORM will use the primary key to delete the record, if primary
> field's blank, GORM will delete all records for the model.

This could totally never bite us in the ass... which is also possibly
representative of my thoughts about gorm in general.

I like many of the decisions and feel of it. You still get a high degree of
expressiveness though, I believe, it will still manifest in a lot of cases as
just embedded SQL which is not great. The ability to configure your DB to
such a large degree through code sets it aside in feel (though maybe not
practice?) from the others. That said the (seeming) inability to do other
things through tags makes for a weird situation where your models are configured
across two places. The configuration itself is also relatively byzantine.

The quality of the documentation is subpar in my opinion leaving me unsure what
I've learned though this experimentation.  Gorm may be great and I just don't
*get it* or it may be a personal work in progress that is mostly there with
shoddy docs and bugs.

The lib is impressive given it's just one person but this also presents
concerns about longevity and community support if we hit a thing. If I'm solo
fixing bugs in an ORM I don't know that I want to be faced with the potential
need to fork it for timely (if any) inclusion.

## <a name="gorp"></a>gorp

gorp is a small library focused on mapping structs into a relational database.
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
gorp seems nice for what it does but the ideal audience feels like it's people
building relatively simple (or extremely complex) services that would like
some sugar around the basics but nothing between them in the actual SQL in
most cases.

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
