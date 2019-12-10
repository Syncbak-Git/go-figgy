//Package figgy provides tags for loading parameters from AWS Parameter Store
package figgy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// maxParameters is the maximum number of parameters that can be requested in a single call to GetParameters
const maxParameters = 10

var durationType reflect.Type = reflect.TypeOf(time.Duration(0))

type Unmarshaler interface {
	UnmarshalParameter(string) error
}

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
	json    bool
	value   reflect.Value
	field   reflect.StructField
}

func newField(key string, decrypt bool) *field {
	return &field{
		key:     strings.TrimSpace(key),
		decrypt: decrypt,
	}
}

// P is a convenience alias for passing paramters to LoadWithParameters
type P map[string]interface{}

// Load AWS Parameter Store parameters based on the defined tags.
//
// When a source type is an array, it is assumed the parameter being loaded
// is a comma separated list.  The list will be split and converted to
// match the array's typing.
//
// You can ignore a field by using "-" for a fields tag.  Unexported fields are also ignored.
func Load(c ssmiface.SSMAPI, v interface{}) (Watcher, error) {
	return LoadWithParameters(c, v, nil)
}

// LoadWithParameters loads AWS Parameter Store parameters based on the defined tags, performing parameter
// substitution on field tags using data-driven templates from "text/template".
//
// When a source type is an array, it is assumed the parameter being loaded
// is a comma separated list.  The list will be split and converted to
// match the array's typing.
//
// You can ignore a field by using "-" for a fields tag.  Unexported fields are also ignored.
func LoadWithParameters(c ssmiface.SSMAPI, v interface{}, data interface{}) (Watcher, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return defaultWatcher{}, &InvalidTypeError{Type: reflect.TypeOf(v)}
	}
	t, err := walk(rv.Elem(), data)
	if err != nil {
		return defaultWatcher{}, err
	}
	if err = load(c, t); err != nil {
		return defaultWatcher{}, err
	}
	if !reflect.ValueOf(data).IsValid() {
		return defaultWatcher{}, err
	}
	return defaultWatcher{
		ssm: c,
		v:   v,
		//make copy of data
		data: reflect.New(reflect.ValueOf(data).Elem().Type()).Interface(),
		hash: hash(data),
	}, err
}

// load fields from AWS Parameter Store
func load(c ssmiface.SSMAPI, f []*field) error {
	plain, decrypt := partitionFields(f, func(x *field) bool {
		return x.decrypt
	})
	err := batchIterateFields(plain, maxParameters, func(f []*field) error {
		return loadParameters(c, f, false)
	})
	if err != nil {
		return err
	}
	return batchIterateFields(decrypt, maxParameters, func(f []*field) error {
		return loadParameters(c, f, true)
	})
}

// in place half stable partition
func partitionFields(f []*field, suffix func(*field) bool) (p1, p2 []*field) {
	var i int
	for ; i < len(f); i++ {
		if suffix(f[i]) {
			break
		}
	}
	for j := i + 1; j < len(f); j++ {
		if !suffix(f[j]) {
			f[i], f[j] = f[j], f[i]
			i++
		}
	}
	return f[:i], f[i:]
}

func batchIterateFields(f []*field, batchSize int, g func([]*field) error) error {
	for i := 0; i < len(f); {
		j := i + batchSize
		if j > len(f) {
			j = len(f)
		}
		if err := g(f[i:j]); err != nil {
			return err
		}
		i = j
	}
	return nil
}

func loadParameters(c ssmiface.SSMAPI, f []*field, decrypt bool) error {
	params, err := getParameters(c, f, decrypt)
	if err != nil {
		return err
	}
	idx := indexParameters(params)
	for _, x := range f {
		p, ok := idx[x.key]
		if !ok {
			return fmt.Errorf("failed to load parameter for key '%s'", x.key)
		}
		err = set(x, aws.StringValue(p.Value))
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

func getParameters(c ssmiface.SSMAPI, f []*field, decrypt bool) ([]*ssm.Parameter, error) {
	res, err := c.GetParameters(&ssm.GetParametersInput{
		Names:          parameterNames(f),
		WithDecryption: aws.Bool(decrypt),
	})
	if err != nil {
		return nil, err
	}
	if len(res.InvalidParameters) != 0 {
		return nil, fmt.Errorf("invalid parameters: %s",
			strings.Join(aws.StringValueSlice(res.InvalidParameters), ", "),
		)
	}
	return res.Parameters, nil
}

func parameterNames(f []*field) []*string {
	names := make([]*string, len(f))
	for i := range f {
		names[i] = aws.String(f[i].key)
	}
	return names
}

func indexParameters(params []*ssm.Parameter) map[string]*ssm.Parameter {
	idx := make(map[string]*ssm.Parameter, len(params))
	for _, p := range params {
		idx[aws.StringValue(p.Name)] = p
	}
	return idx
}

// walk the value recursively to initialize pointers and build a graph of fields and tag options
func walk(v reflect.Value, data interface{}) ([]*field, error) {
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
		pf, err := tag(ft, data)
		if err != nil {
			return nil, err
		}
		if pf != nil {
			pf.field = ft
			pf.value = fv
			p = append(p, pf)
		} else {
			// only walk down embedded structs with no 'ssm' tag
			switch fv.Kind() {
			case reflect.Struct:
				tags, err := walk(fv, data)
				if err != nil {
					return nil, err
				}
				p = append(p, tags...)
				continue
			}
		}
	}
	return p, nil
}

// tag parses the ssm tag from a given field
func tag(f reflect.StructField, data interface{}) (*field, error) {
	t := f.Tag.Get("ssm")
	if t == "" || t == "-" {
		return nil, nil
	}
	o := strings.Split(t, ",")
	fld := newField(strings.TrimSpace(o[0]), false)
	if fld.key == "" {
		return nil, &TagParseError{Tag: t, Field: f.Name}
	}
	tpl, err := template.New(fld.key).Parse(fld.key)
	if err == nil {
		b := &bytes.Buffer{}
		err = tpl.Execute(b, data)
		if err == nil {
			fld.key = b.String()
		}
	}
	for _, option := range o[1:] {
		switch strings.TrimSpace(option) {
		case "decrypt":
			fld.decrypt = true
		case "json":
			fld.json = true
		}
	}
	return fld, nil
}

// set will attempt to set the underlying value based on the value's type
func set(f *field, s string) error {
	v := f.value
	if !v.CanSet() {
		return errors.New(v.Type().String() + " cannot be set")
	}
	if u := unmarshaler(v); u != nil {
		if f.json {
			return fmt.Errorf("cannot use 'json' option on a type with a custom unmarshaller: %s %s", f.field.Name, f.field.Type.String())
		}
		return u.UnmarshalParameter(s)
	}
	if f.json {
		return setJSON(f, s)
	}
	// special case with time.Duration and assignable types
	if v.Type().AssignableTo(durationType) {
		if p, err := time.ParseDuration(s); err == nil {
			v.Set(reflect.ValueOf(p))
			return nil
		}
	}
	switch v.Kind() {
	// handles the case data types are wrapped in other constructs, EG slices
	case reflect.Ptr:
		// create new pointer to a zero value
		new := reflect.New(v.Type().Elem())
		set(&field{value: new.Elem()}, s)
		// assign new pointer
		v.Set(new)
		break
	case reflect.Slice:
		// we assume the list is separated by commas
		l := strings.Split(s, ",")
		sz := len(l)
		v.Set(reflect.MakeSlice(v.Type(), sz, sz))
		for i, w := range l {
			set(&field{value: v.Index(i)}, w)
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

func unmarshaler(v reflect.Value) Unmarshaler {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		if u, ok := v.Interface().(Unmarshaler); ok {
			return u
		}
	}
	return nil
}

func setJSON(f *field, s string) error {
	v := f.value
	if v.Kind() != reflect.Ptr {
		if !v.CanAddr() {
			return fmt.Errorf("%s is not addressable", v.Type().String())
		}
		v = v.Addr()
	}
	if !v.CanInterface() {
		return fmt.Errorf("%s is not interfaceable", v.Type().String())
	}
	if err := json.Unmarshal([]byte(s), v.Interface()); err != nil {
		return fmt.Errorf("json unmarshal error for field '%s'", f.field.Name)
	}
	return nil
}
