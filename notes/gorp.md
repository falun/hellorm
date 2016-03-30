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
