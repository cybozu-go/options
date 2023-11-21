package options

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// Option[T] represents an optional value of type T.
//
// Options that have no value are called [None].
// The zero value of Option[T] is None.
type Option[T any] struct {
	// invariant: !present => value == <zero value of T>
	value   T
	present bool
}

// New returns a new Option[T] with the given value.
func New[T any](value T) Option[T] {
	return Option[T]{
		value:   value,
		present: true,
	}
}

// None returns a new Option[T] with no value.
func None[T any]() Option[T] {
	return Option[T]{}
}

// FromPointer creates Option[T] from a pointer.
// If the pointer is nil, None is returned.
// Otherwise, a new Option[T] with the pointed value is returned.
func FromPointer[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	} else {
		return New(*ptr)
	}
}

// FromTuple creates Option[T] from a tuple of (T, bool).
// If the bool is true, a new Option[T] with the given value is returned.
// Otherwise, None is returned.
func FromTuple[T any](value T, present bool) Option[T] {
	if present {
		return New(value)
	} else {
		return None[T]()
	}
}

// IsPresent returns true if the option has a value.
func (o *Option[T]) IsPresent() bool {
	return o.present
}

// IsNone returns true if the option is None.
func (o *Option[T]) IsNone() bool {
	return !o.present
}

// Unwrap returns the value of the option.
// If the option is None, Unwrap panics.
// You should check the option with [Option.IsPresent] before calling this method.
func (o *Option[T]) Unwrap() T {
	if o.present {
		return o.value
	} else {
		panic(fmt.Errorf("Option[%T].Unwrap: unwrapping None value", o.value))
	}
}

// UnwrapOr returns the value of the option.
// If the option is None, the given default value is returned.
func (o *Option[T]) UnwrapOr(defaultValue T) T {
	if o.present {
		return o.value
	} else {
		return defaultValue
	}
}

// UnwrapOrZero returns the value of the option.
// If the option is None, the zero value of T is returned.
func (o *Option[T]) UnwrapOrZero() T {
	return o.value
}

// Pointer returns a pointer to the wrapped value of the option.
// If the option is None, nil is returned.
func (o *Option[T]) Pointer() *T {
	if o.present {
		return &o.value
	} else {
		return nil
	}
}

// Map returns a new option by applying the given function to the value of the option.
// If the option is None, None is returned.
func Map[A any, B any](o Option[A], f func(A) B) Option[B] {
	if o.present {
		return New(f(o.value))
	} else {
		return None[B]()
	}
}

// String returns the string representation of the wrapped value.
// If the option is None, an empty string is returned.
func (o Option[T]) String() string {
	if o.present {
		return fmt.Sprint(o.value)
	} else {
		return ""
	}
}

// GoString returns the Go representation of the option.
func (o Option[T]) GoString() string {
	if o.present {
		return fmt.Sprintf("options.New(%#v)", o.value)
	} else {
		return fmt.Sprintf("options.None[%T]()", o.value)
	}
}

// MarshalJSON implements the [json.Marshaler] interface.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Pointer())
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (o *Option[T]) UnmarshalJSON(bytes []byte) error {
	var p *T
	if err := json.Unmarshal(bytes, &p); err != nil {
		return fmt.Errorf("Option[%T].UnmarshalJSON: %w", o.value, err)
	}
	*o = FromPointer(p)
	return nil
}

// Value implements the SQL [driver.Valuer] interface.
// See http://jmoiron.net/blog/built-in-interfaces
func (o Option[T]) Value() (driver.Value, error) {
	if o.present {
		return o.value, nil
	} else {
		return nil, nil
	}
}

// Scan implements the SQL [driver.Scanner] interface.
// See http://jmoiron.net/blog/built-in-interfaces
func (o *Option[T]) Scan(src any) error {
	if src == nil {
		*o = None[T]()
		return nil
	}
	av, err := driver.DefaultParameterConverter.ConvertValue(src)
	if err != nil {
		return fmt.Errorf("Option[%T].Scan: failed to convert value from SQL driver: %w", o.value, err)
	}
	v, ok := av.(T)
	if !ok {
		return fmt.Errorf("Option[%T].Scan: failed to convert value %#v to type %T", o.value, av, o.value)
	}
	*o = New(v)
	return nil
}

// Equal returns true if the two options are equal.
// Equality of the wrapped values is determined by [reflect.DeepEqual].
//
// Usually you don't need to call this method since you can use == operator.
// This method is provided to make Option[T] comparable by [go-cmp].
//
// [go-cmp]: https://github.com/google/go-cmp
func (o Option[T]) Equal(other Option[T]) bool {
	if o.present != other.present {
		return false
	}
	if !o.present {
		return true
	}
	return reflect.DeepEqual(o.value, other.value)
}

// Pointer is a free function version of [Option.Pointer].
//
// This function is provided to write Transfermer of [go-cmp].
// See README for details.
//
// [go-cmp]: https://github.com/google/go-cmp
func Pointer[T any](o Option[T]) *T {
	return o.Pointer()
}
