### hellorm

Problem: I need to talk to a database and hate the following things:

- having the same code scattered across 84 bits of source
- panics at runtime
- `interface{}` and loss of type safety


Soooo I'm evaluating some ORMs and collect my thoughts about them. If you're
looking at this don't take this to be super informed, I've got like 3 months
of go behind me at the time of this survey.

### The libraries:

* [ryo](./notes/ryo.md) - Completed enough; code under `./ryo`
* [gorm](./notes/gorm.md) - Completed; code under `./gm`
* [gorp](./notes/gorp.md) - Completed; code under `./g`
* [beego/orm](./notes/beego.md) - Completed; code under `./b`
