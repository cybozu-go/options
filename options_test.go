package options_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cybozu-go/options"
)

func assertEqual[T comparable](t *testing.T, a, b T) {
	t.Helper()
	if a != b {
		t.Errorf("not equal: a='%#v', b='%#v'", a, b)
	}
}

func assertDeepEqual[T any](t *testing.T, a, b T) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("not equal: a='%#v', b='%#v'", a, b)
	}
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	j, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(j)
}

func unmarshal[T any](t *testing.T, j string) *T {
	var v T
	err := json.Unmarshal([]byte(j), &v)
	if err != nil {
		t.Fatal(err)
	}
	return &v
}

func toSQLValue[T any](t *testing.T, opt options.Option[T]) driver.Value {
	value, err := driver.DefaultParameterConverter.ConvertValue(opt)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func ExampleFromTuple() {
	some := options.FromTuple(42, true)
	fmt.Println(some.GoString())

	none := options.FromTuple[int](0, false)
	fmt.Println(none.GoString())

	// Output:
	// options.New(42)
	// options.None[int]()
}

func ExampleOption_Unwrap() {
	opt := options.New(42)
	fmt.Println(opt.Unwrap())
	// Output:
	// 42
}

func ExampleOption_UnwrapOr() {
	some := options.New(42)
	fmt.Println(some.UnwrapOr(-1))

	none := options.None[int]()
	fmt.Println(none.UnwrapOr(-1))

	// Output:
	// 42
	// -1
}

func ExampleOption_UnwrapOrZero() {
	some := options.New(42)
	fmt.Println(some.UnwrapOrZero())

	none := options.None[int]()
	fmt.Println(none.UnwrapOrZero())

	// Output:
	// 42
	// 0
}

func ExampleMap() {
	getLength := func(s string) int { return len(s) }

	some := options.New("hello")
	fmt.Printf("some: %#v\n", options.Map(some, getLength))

	none := options.None[string]()
	fmt.Printf("none: %#v\n", options.Map(none, getLength))

	// Output:
	// some: options.New(5)
	// none: options.None[int]()
}

func ExampleOption_String() {
	some := options.New(true)
	fmt.Println("some:", some.String())

	none := options.None[bool]()
	fmt.Println("none:", none.String())

	// Output:
	// some: true
	// none:
}

func ExampleOption_GoString() {
	some := options.New(true)
	fmt.Printf("some: %#v\n", some)

	none := options.None[bool]()
	fmt.Printf("none: %#v\n", none)

	// Output:
	// some: options.New(true)
	// none: options.None[bool]()
}

func TestJSONMarshal(t *testing.T) {
	opt1 := options.New(3.14)
	assertEqual(t, marshal(t, opt1), `3.14`)

	opt2 := options.New("hello")
	assertEqual(t, marshal(t, opt2), `"hello"`)

	opt3 := options.None[string]()
	assertEqual(t, marshal(t, opt3), `null`)

	ts, err := time.Parse(time.RFC3339, "2021-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	opt4 := options.New(ts)
	assertEqual(t, marshal(t, opt4), `"2021-02-03T04:05:06Z"`)

	opt5 := options.New([]string{"foo", "bar"})
	assertEqual(t, marshal(t, opt5), `["foo","bar"]`)

	opt6 := options.New(map[string]int{"foo": 1, "bar": 2})
	assertEqual(t, marshal(t, opt6), `{"bar":2,"foo":1}`)
}

func TestJSONUnmarshal(t *testing.T) {
	json1 := `3.14`
	opt1 := unmarshal[options.Option[float64]](t, json1)
	assertEqual(t, *opt1, options.New(3.14))

	json2 := `"hello"`
	opt2 := unmarshal[options.Option[string]](t, json2)
	assertEqual(t, *opt2, options.New("hello"))

	json3 := `null`
	opt3 := unmarshal[options.Option[string]](t, json3)
	assertEqual(t, *opt3, options.None[string]())

	json4 := `"2021-02-03T04:05:06Z"`
	opt4 := unmarshal[options.Option[time.Time]](t, json4)
	ts, err := time.Parse(time.RFC3339, "2021-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, *opt4, options.New(ts))

	json5 := `["foo","bar"]`
	opt5 := unmarshal[options.Option[[]string]](t, json5)
	assertDeepEqual(t, *opt5, options.New([]string{"foo", "bar"}))

	json6 := `{"bar":2,"foo":1}`
	opt6 := unmarshal[options.Option[map[string]int]](t, json6)
	assertDeepEqual(t, *opt6, options.New(map[string]int{"foo": 1, "bar": 2}))
}

func TestSQLValue(t *testing.T) {
	opt1 := options.New(3.14)
	value1 := toSQLValue(t, opt1)
	assertEqual[any](t, value1, 3.14)

	opt2 := options.None[float64]()
	value2 := toSQLValue(t, opt2)
	assertEqual[any](t, value2, nil)

	opt3 := options.New("hello")
	value3 := toSQLValue(t, opt3)
	assertEqual[any](t, value3, "hello")

	ts := time.Now()
	opt4 := options.New[time.Time](ts)
	value4 := toSQLValue(t, opt4)
	assertEqual[any](t, value4, ts)

	opt5 := options.None[time.Time]()
	value5 := toSQLValue(t, opt5)
	assertEqual[any](t, value5, nil)
}

func TestSQLScan(t *testing.T) {
	nullString1, _ := sql.NullString{String: "hello", Valid: true}.Value()
	var opt1 options.Option[string]
	if err := opt1.Scan(nullString1); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, opt1, options.New("hello"))

	nullString2, _ := sql.NullString{String: "", Valid: false}.Value()
	var opt2 options.Option[string]
	if err := opt2.Scan(nullString2); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, opt2, options.None[string]())

	ts := time.Now()
	nullTime, _ := sql.NullTime{Time: ts, Valid: true}.Value()
	var opt3 options.Option[time.Time]
	if err := opt3.Scan(nullTime); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, opt3, options.New[time.Time](ts))
}

func TestEqual(t *testing.T) {
	assertEqual(t, options.New(3.14).Equal(options.New(3.14)), true)
	assertEqual(t, options.New(3.14).Equal(options.New(1.59)), false)
	assertEqual(t, options.New(3.14).Equal(options.None[float64]()), false)
	assertEqual(t, options.None[float64]().Equal(options.None[float64]()), true)
	assertEqual(t, options.None[float64]().Equal(options.New(3.14)), false)
	assertEqual(t, options.New("hello").Equal(options.New("hello")), true)
}
