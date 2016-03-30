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

## sqlx Addendum

It's not really an ORM but some sugar for `database/sql` ([github](https://github.com/jmoiron/sqlx))
and nifty helpers to scan into structs. Bonus points in that jmoiron hangs out
in `#go-nuts` and seems generally helpful.

I expect I'll fold this into `ryo` if it goes to prod if I can figure out
logging as I have no desire to rewrap everything, again.
