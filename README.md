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

## Compare complex values by go-cmp

Although `Option[T]` can be compared by go-cmp, it has some caveats when `T` is complex.

- `cmp.Diff` may be hard to read because diff of `Option[T]` is shown as diff of `Option[T].String()`.
- `cmp.Diff` does not compare unexported fields of structs by default. On the other hand, `Option[T].Equal` is based on `reflect.DeepEqual`, which compares unexported fields.

You can use `cmp.Transformer` in such case.

```go
type NestedData struct {
	Value string
}

type TestData struct {
	Value  string
	Nested *NestedData
}

func TestGoCmp(t *testing.T) {
	// Use *T instead of Option[T] in the cmp.Diff.
	cmpopt := cmp.Transformer("options.Option", options.Pointer[*TestData])

	d1 := options.New(&TestData{
		Value: "test",
		Nested: &NestedData{
			Value: "test",
		},
	})
	d2 := options.New(&TestData{
		Value: "test",
		Nested: &NestedData{
			Value: "test2",
		},
	})

	if diff := cmp.Diff(d1, d2, cmpopt); diff != "" {
		t.Errorf("diff:\n%s", diff)
	}
}
```
