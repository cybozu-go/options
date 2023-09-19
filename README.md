# options

`cybozu-go/options` provides `Option[T]`, which represents an optional value of type `T`.

## Example

```go
opt := options.New(42) // or option.None[int]()

fmt.Println(opt) // prints "42"
fmt.Printf("%#v\n", opt) // prints "options.New(42)"

if opt.IsPresent() {
    // opt.Unwrap panics when opt is None.
    // When there are feasible default values, you can use UnwrapOr or UnwrapOrZero, which do not panic.
    v := opt.Unwrap()
    DoSomething(v)
}
```

## Interoperability

- `Option[T]` can be serialized into or deserialized from JSON by `encoding/json`.
    - An `Option[T]` is serialized as if it is `*T`.
- `Option[T]` can be inserted into or selected from databases by `database/sql`.
    - `Option[string]` is handled as if it is `sql.NullString`, `Option[time.Time]` is handled as if it is `sql.NullTime`, and so on.
- `Option[T]` can be compared by [google/go-cmp](https://github.com/google/go-cmp).
    - `Option[T].Equal` is implemented sololy for this purpose.
