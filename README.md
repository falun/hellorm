### hellorm

Problem: I need to talk to a database and hate the following things:

- having the same code scattered across 84 bits of source
- panics at runtime
- `interface{}` and loss of type safety


Soooo I'm evaluating some ORMs and collect my thoughts about them. If you're
looking at this don't take this to be super informed, I've got like 2 months
of go behind me at the time of this survey.

### The libraries:

* [NIH](./notes.md#nih) - :sob:, just wirte something from scratch (TBD)
* [gorm](./notes.md#gorm) - Completed; code under `./gm`
* [gorp](./notes.md#gorp) - Completed; code under `./g`
* [beego/orm](./notes.md#beego) - Completed; code under `./b`
