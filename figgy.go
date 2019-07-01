//Package figgy provides tags for loading parameters from AWS Parameter Store
package figgy

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// InvalidTypeError descibes an invalid argument passed to Load.
type InvalidTypeError struct {
	Type reflect.Type
}

func (e *InvalidTypeError) Error() string {
	if e.Type == nil {
		return "nil type"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "non-pointer Load(" + e.Type.String() + ")"
	}
	return "invalid type " + e.Type.String()
}

// TagParseError describes an invalid tag
type TagParseError struct {
	// Tag that failed to be fully parsed
	Tag string
	// Field metadata that the tag is parsed from
	Field string
}

func (e *TagParseError) Error() string {
	return "failed to parse tag [" + e.Tag + "] for field " + e.Field
}

// ConvertTypeError describes a value that failed to be set for a field
type ConvertTypeError struct {
	//Field that the value was being assigned to
	Field string
	// Type of value that couldn't be assigned
	Type string
	// Value that failed to be converted
	Value string
}

func (e *ConvertTypeError) Error() string {
	if e.Field != "" || e.Type != "" {
		return "failed to convert '" + e.Value + "' to " + e.Type + " for field " + e.Field
	}
	return "failed to convert '" + e.Value + "'"
}

// field represents parse struct fields tags and the underlying value
type field struct {
	key     string
	decrypt bool
	value   reflect.Value
	field   reflect.StructField
}

func newField(key string, decrypt bool) *field {
	return &field{
		key:     strings.TrimSpace(key),
		decrypt: decrypt,
	}
}

// Load AWS Parameter Store parameters based on the defined tags.
//
// When a source type is an array, it is assumed the parameter being loaded
// is a comma seperated list.  The list will be split and converted to
// match the array's typing.
//
// You can ignore a field by using "-" for a fields tag.  Unexported fields are also ignored.
func Load(c ssmiface.SSMAPI, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidTypeError{Type: reflect.TypeOf(v)}
	}
	t, err := walk(rv.Elem())
	if err != nil {
		return err
	}
	return load(c, t)
}

// load fields from AWS Parameter Store
func load(c ssmiface.SSMAPI, f []*field) error {
	for _, x := range f {
		//TODO: Group parameters to minimize calls
		p := &ssm.GetParameterInput{
			Name:           aws.String(x.key),
			WithDecryption: aws.Bool(x.decrypt),
		}
		out, err := c.GetParameter(p)
		if err != nil {
			return err
		}
		err = set(x.value, *out.Parameter.Value)
		if err != nil {
			switch err := err.(type) {
			case *ConvertTypeError:
				//enrich the error with the field
				err.Field = x.field.Name
				return err
			}
			return err
		}
	}
	return nil
}

// walk the value recursively to initialize pointers and build a graph of fields and tag options
func walk(v reflect.Value) ([]*field, error) {
	p := make([]*field, 0)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		// ignore unexported field
		if ft.PkgPath != "" {
			continue
		}
		// handles initializing a ptr and gets the underlying value to operate on
		if fv.Kind() == reflect.Ptr {
			fv.Set(reflect.New(fv.Type().Elem()))
			fv = reflect.Indirect(fv)
		}
		switch fv.Kind() {
		case reflect.Struct:
			tags, err := walk(fv)
			if err != nil {
				return nil, err
			}
			p = append(p, tags...)
			continue
		}
		pf, err := tag(ft)
		if err != nil {
			return nil, err
		}
		if pf != nil {
			pf.field = ft
			pf.value = fv
			p = append(p, pf)
		}
	}
	return p, nil
}

// tag parses the ssm tag from a given field
func tag(f reflect.StructField) (*field, error) {
	t := f.Tag.Get("ssm")
	if t == "" || t == "-" {
		return nil, nil
	}
	o := strings.Split(t, ",")
	fld := newField(strings.TrimSpace(o[0]), false)
	if fld.key == "" {
		return nil, &TagParseError{Tag: t, Field: f.Name}
	}
	for _, option := range o[1:] {
		switch strings.TrimSpace(option) {
		case "decrypt":
			fld.decrypt = true
		}
	}
	return fld, nil
}

// set will attempt to set the underlying value based on the value's type
func set(v reflect.Value, s string) error {
	if !v.CanSet() {
		return errors.New(v.Type().String() + " cannot be set")
	}
	switch v.Kind() {
	// handles the case data types are wrapped in other constructs, EG slices
	case reflect.Ptr:
		// create new pointer to a zero value
		new := reflect.New(v.Type().Elem())
		set(new.Elem(), s)
		// assign new pointer
		v.Set(new)
		break
	case reflect.Slice:
		// we assume the list is seperated by commas
		l := strings.Split(s, ",")
		sz := len(l)
		v.Set(reflect.MakeSlice(v.Type(), sz, sz))
		for i, w := range l {
			set(v.Index(i), w)
		}
		break
	case reflect.String:
		v.SetString(s)
		break
	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return &ConvertTypeError{
				Type:  v.Type().String(),
				Value: s,
			}
		}
		v.SetBool(n)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			return &ConvertTypeError{
				Type:  v.Type().String(),
				Value: s,
			}
		}
		v.SetInt(n)
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			return &ConvertTypeError{
				Type:  v.Type().String(),
				Value: s,
			}
		}
		v.SetUint(n)
		break
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			return &ConvertTypeError{
				Type:  v.Type().String(),
				Value: s,
			}
		}
		v.SetFloat(n)
		break
	}
	return nil
}
