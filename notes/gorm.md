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
