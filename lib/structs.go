package shush

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal calls UnmarshalContext with a default context with no timeout.
// You should use UnmarshalContext if possible.
func (s *Session) Unmarshal(o interface{}) error {
	ctx := context.Background()

	return s.UnmarshalContext(ctx, o)
}

// UnmarshalContext unmarshals a struct decorated with `shush` struct tags and
// fetches and populates each field
func (s *Session) UnmarshalContext(ctx context.Context, o interface{}) error {
	rv := reflect.ValueOf(o)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("expected input struct should be a pointer")
	}

	v := rv.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		tag, ok := f.Tag.Lookup("shush")
		if !ok {
			return errors.New("missing shush tag")
		}

		if tag == "" {
			return errors.New("shush struct tag is empty")
		}

		if !fv.CanSet() {
			return fmt.Errorf("cannot set capstone value at field %s", f.Name)
		}

		value, _, err := s.Get(ctx, tag)
		if err != nil {
			return err
		}

		switch f.Type.Kind() {
		case reflect.String:
			fv.SetString(value)
		default:
			return errors.New("unusable destination field type")
		}
	}

	return nil
}
