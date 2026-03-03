package figgy

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	Data      map[string]string
	batchSize int
}

func (c *MockClient) GetValues(keys []string) (map[string]string, error) {
	if c.batchSize > 0 && len(keys) > c.batchSize {
		return nil, fmt.Errorf("max parameters exceeded: received %d, max %d", len(keys), c.batchSize)
	}
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v, ok := c.Data[k]
		if !ok {
			return nil, fmt.Errorf("invalid parameters: %s", k)
		}
		out[k] = v
	}
	return out, nil
}

func (c *MockClient) BatchSize() int {
	if c.batchSize > 0 {
		return c.batchSize
	}
	return 10
}

func NewMockClient() *MockClient {
	return &MockClient{
		batchSize: 10,
		Data: map[string]string{
			"bool":             "true",
			"int":              "2",
			"int8":             "3",
			"int16":            "4",
			"int32":            "5",
			"int64":            "6",
			"uint":             "7",
			"uint8":            "8",
			"uint16":           "9",
			"uint32":           "10",
			"uint64":           "11",
			"uintptr":          "12",
			"float32":          "12.1",
			"float64":          "12.2",
			"duration":         "3600000000000",
			"durationstring":   "3600s",
			"pbool":            "true",
			"pint":             "13",
			"pint8":            "14",
			"pint16":           "15",
			"pint32":           "16",
			"pint64":           "17",
			"puint":            "18",
			"puint8":           "19",
			"puint16":          "20",
			"puint32":          "21",
			"puint64":          "22",
			"puintptr":         "23",
			"pfloat32":         "23.1",
			"pfloat64":         "23.2",
			"string":           "this is a string",
			"pstring":          "this is a ptr to a string",
			"sliceint":         "1,2,3,4,5",
			"pduration":        "3600000000000",
			"pdurationstring":  "3600s",
			"simplejson":       `{"F1": 1, "F2": "2"}`,
			"simplejsonarray":  `[{"F1": 1, "F2": "2"}]`,
			"badjson":          "invalid",
		},
	}
}

func NewTypes() *Types {
	return &Types{
		unexported: 100,
	}
}

type Types struct {
	Bool           bool          `ssm:"bool"`
	Int            int           `ssm:"int"`
	Int8           int8          `ssm:"int8"`
	Int16          int16         `ssm:"int16"`
	Int32          int32         `ssm:"int32"`
	Int64          int64         `ssm:"int64"`
	Uint           uint          `ssm:"uint"`
	Uint8          uint8         `ssm:"uint8"`
	Uint16         uint16        `ssm:"uint16"`
	Uint32         uint32        `ssm:"uint32"`
	Uint64         uint64        `ssm:"uint64"`
	Uintptr        uintptr       `ssm:"uintptr"`
	Float32        float32       `ssm:"float32"`
	Float64        float64       `ssm:"float64"`
	Duration       time.Duration `ssm:"duration"`
	DurationString time.Duration `ssm:"durationstring"`

	PBool    *bool    `ssm:"pbool"`
	PInt     *int     `ssm:"pint"`
	PInt8    *int8    `ssm:"pint8"`
	PInt16   *int16   `ssm:"pint16"`
	PInt32   *int32   `ssm:"pint32"`
	PInt64   *int64   `ssm:"pint64"`
	PUint    *uint    `ssm:"puint"`
	PUint8   *uint8   `ssm:"puint8"`
	PUint16  *uint16  `ssm:"puint16"`
	PUint32  *uint32  `ssm:"puint32"`
	PUint64  *uint64  `ssm:"puint64"`
	PUintptr *uintptr `ssm:"puintptr"`
	PFloat32 *float32 `ssm:"pfloat32"`
	PFloat64 *float64 `ssm:"pfloat64"`

	String  string  `ssm:"string"`
	PString *string `ssm:"pstring"`

	Slice  []int  `ssm:"sliceint"`
	SliceP []*int `ssm:"sliceint"`

	Nested  Nested
	PNested *Nested

	Top  Top
	PTop *Top

	unexported int
}

type str string

func (c *str) UnmarshalParameter(s string) error {
	*c = str("cs-" + s)
	return nil
}

type Nested struct {
	String  string  `ssm:"string"`
	PString *string `ssm:"pstring"`
}

type Top struct {
	String  string  `ssm:"string"`
	PString *string `ssm:"pstring"`
	Nested  Nested
}

func TestNonPtrAndNilInput(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want error
	}{
		"nil":     {in: nil, want: &InvalidTypeError{Type: nil}},
		"non ptr": {in: struct{}{}, want: &InvalidTypeError{Type: reflect.TypeOf(struct{}{})}},
	}

	m := NewMockClient()
	for n, tc := range tests {
		err := Load(m, tc.in)
		assert.EqualErrorf(t, err, tc.want.Error(), "unexpected error while executing test %s", n)
	}
}

func TestTypeConvert(t *testing.T) {
	ex := NewTypes()
	err := Load(NewMockClient(), ex)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all values were correctly populated
	assert.Equal(t, true, ex.Bool)
	assert.Equal(t, 2, ex.Int)
	assert.Equal(t, int8(3), ex.Int8)
	assert.Equal(t, int16(4), ex.Int16)
	assert.Equal(t, int32(5), ex.Int32)
	assert.Equal(t, int64(6), ex.Int64)
	assert.Equal(t, uint(7), ex.Uint)
	assert.Equal(t, uint8(8), ex.Uint8)
	assert.Equal(t, uint16(9), ex.Uint16)
	assert.Equal(t, uint32(10), ex.Uint32)
	assert.Equal(t, uint64(11), ex.Uint64)
	assert.Equal(t, uintptr(12), ex.Uintptr)
	assert.Equal(t, float32(12.1), ex.Float32)
	assert.Equal(t, 12.2, ex.Float64)
	assert.Equal(t, "this is a string", ex.String)
	assert.Equal(t, "this is a ptr to a string", *ex.PString)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, ex.Slice)

	// Duration from nanosecond integer (falls through ParseDuration to int64 parser)
	assert.Equal(t, time.Hour, ex.Duration)
	// Duration from string format
	assert.Equal(t, time.Hour, ex.DurationString)

	// Pointer types
	assert.NotNil(t, ex.PBool)
	assert.Equal(t, true, *ex.PBool)
	assert.NotNil(t, ex.PInt)
	assert.Equal(t, 13, *ex.PInt)
	assert.NotNil(t, ex.PFloat64)
	assert.Equal(t, 23.2, *ex.PFloat64)
}

func TestUnmarshalIface(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want interface{}
	}{
		"unmarshal string type alias": {
			in: &struct {
				AliasString str `ssm:"string"`
			}{},
			want: &struct {
				AliasString str
			}{
				AliasString: "cs-this is a string",
			},
		}}

	for n, tc := range tests {
		err := Load(NewMockClient(), tc.in)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(t, tc.in, tc.want, "test '%s' failed", n)
	}
}

func TestTypeConvertErrors(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want error
	}{
		"invalid bool convert": {in: &struct {
			Bool bool `ssm:"string"`
		}{}, want: &ConvertTypeError{Field: "Bool", Type: "bool", Value: "this is a string"}},
		"invalid int convert": {in: &struct {
			Int int `ssm:"string"`
		}{}, want: &ConvertTypeError{Field: "Int", Type: "int", Value: "this is a string"}},
		"invalid uint convert": {in: &struct {
			UInt uint `ssm:"string"`
		}{}, want: &ConvertTypeError{Field: "UInt", Type: "uint", Value: "this is a string"}},
		"invalid float convert": {in: &struct {
			Float32 float32 `ssm:"string"`
		}{}, want: &ConvertTypeError{Field: "Float32", Type: "float32", Value: "this is a string"}},
		"invalid time.Duration convert": {in: &struct {
			Duration time.Duration `ssm:"string"`
		}{}, want: &ConvertTypeError{Field: "Duration", Type: "time.Duration", Value: "this is a string"}},
	}

	for n, tc := range tests {
		err := Load(NewMockClient(), tc.in)
		assert.EqualError(t, err, tc.want.Error(), "test '%s' failed", n)
	}
}

func TestInvalidParams(t *testing.T) {
	var c struct {
		Invalid string `ssm:"/no/such/param"`
	}
	err := Load(NewMockClient(), &c)
	assert.Error(t, err)
}

func TestMixedPlainAndDecryptParams(t *testing.T) {
	var c struct {
		Plain1   string `ssm:"string"`
		Plain2   bool   `ssm:"bool"`
		Decrypt1 int    `ssm:"int,decrypt"`
		Decrypt2 int32  `ssm:"int32,decrypt"`
	}
	err := Load(NewMockClient(), &c)
	assert.NoError(t, err)
	assert.Equal(t, c.Plain1, "this is a string")
	assert.Equal(t, c.Plain2, true)
	assert.Equal(t, c.Decrypt1, 2)
	assert.Equal(t, c.Decrypt2, int32(5))
}

type JSONTest struct {
	JSON  SimpleJSON  `ssm:"simplejson,json"`
	PJSON *SimpleJSON `ssm:"simplejson,json"`
	EJSON struct {
		F1 int
		F2 string
	} `ssm:"simplejson,json"`
	AJSON []SimpleJSON `ssm:"simplejsonarray,json"`
}

type SimpleJSON struct {
	F1 int
	F2 string
}

func TestJSON(t *testing.T) {
	var j JSONTest
	err := Load(NewMockClient(), &j)
	assert.NoError(t, err)
	s := SimpleJSON{F1: 1, F2: "2"}
	assert.Equal(t, s, j.JSON)
	assert.NotNil(t, j.PJSON)
	assert.Equal(t, s, *j.PJSON)
	assert.EqualValues(t, s, j.EJSON)
	assert.Len(t, j.AJSON, 1)
	assert.Equal(t, s, j.AJSON[0])
}

func TestJSONError(t *testing.T) {
	var j struct {
		SimpleJSON `ssm:"badjson,json"`
	}
	err := Load(NewMockClient(), &j)
	assert.Error(t, err)
}

func TestJSONWithUnmarshallerError(t *testing.T) {
	var j struct {
		Test str `ssm:"string,json"`
	}
	err := Load(NewMockClient(), &j)
	assert.Error(t, err)
}

func TestTagParse(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		data interface{}
		want *field
		err  error
	}{
		"key only": {in: struct {
			Field string `ssm:"parsed"`
		}{}, want: &field{key: "parsed"}, err: nil},
		"with decrypt": {in: struct {
			Field string `ssm:"parsed,decrypt"`
		}{}, want: &field{key: "parsed", decrypt: true}, err: nil},
		"without key": {in: struct {
			Field string `ssm:","`
		}{}, want: nil, err: &TagParseError{Tag: ",", Field: "Field"}},
		"empty tag": {in: struct {
			Field string `ssm:""`
		}{}, want: nil, err: nil},
		"ignoreKey": {in: struct {
			Field string `ssm:"-"`
		}{}, want: nil, err: nil},
		"with parameter": {in: struct {
			Fields string `ssm:"/{{.env}}/environment"`
		}{}, want: &field{key: "/dev/environment"},
			data: P{"env": "dev"}},
		"with json": {in: struct {
			Field string `ssm:"simplejson,json"`
		}{}, want: &field{key: "simplejson", json: true}, err: nil},
	}

	for n, tc := range tests {
		f := reflect.TypeOf(tc.in).Field(0) //Not the safest assumption
		tag, err := tag(f, tc.data)
		if tc.want != nil {
			assert.Equalf(t, tc.want.key, tag.key, "keys are do not match for test %s", n)
			assert.Equalf(t, tc.want.decrypt, tag.decrypt, "decrypt flag does not match for test %s", n)
		}
		if err != nil {
			assert.EqualError(t, err, tc.err.Error())
		}
	}
}

func TestBatchIterate(t *testing.T) {
	fields := make([]*field, 25)
	for i := range fields {
		fields[i] = &field{key: fmt.Sprintf("key%d", i)}
	}
	var batches int
	err := batchIterateFields(fields, 10, func(f []*field) error {
		batches++
		if batches < 3 {
			assert.Len(t, f, 10)
		} else {
			assert.Len(t, f, 5)
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, batches)
}

func TestLoadWithParameters(t *testing.T) {
	m := NewMockClient()
	m.Data["/dev/environment"] = "development"

	var c struct {
		Env string `ssm:"/{{.env}}/environment"`
	}
	err := LoadWithParameters(m, &c, P{"env": "dev"})
	assert.NoError(t, err)
	assert.Equal(t, "development", c.Env)
}

func TestLoadEmptyStruct(t *testing.T) {
	m := NewMockClient()
	var c struct {
		Ignored string
	}
	err := Load(m, &c)
	assert.NoError(t, err)
}

func TestLoadBatchSizeOne(t *testing.T) {
	m := &MockClient{
		batchSize: 1,
		Data: map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}
	var c struct {
		A string `ssm:"a"`
		B string `ssm:"b"`
		C string `ssm:"c"`
	}
	err := Load(m, &c)
	assert.NoError(t, err)
	assert.Equal(t, "1", c.A)
	assert.Equal(t, "2", c.B)
	assert.Equal(t, "3", c.C)
}

func TestLoadExactBatchSize(t *testing.T) {
	m := &MockClient{
		batchSize: 2,
		Data: map[string]string{
			"a": "1",
			"b": "2",
		},
	}
	var c struct {
		A string `ssm:"a"`
		B string `ssm:"b"`
	}
	err := Load(m, &c)
	assert.NoError(t, err)
	assert.Equal(t, "1", c.A)
	assert.Equal(t, "2", c.B)
}

func TestLoadGetValuesError(t *testing.T) {
	m := &MockClient{
		batchSize: 10,
		Data:      map[string]string{}, // empty — all keys will fail
	}
	var c struct {
		Missing string `ssm:"nope"`
	}
	err := Load(m, &c)
	assert.Error(t, err)
}

// partialMockClient returns a map missing some requested keys (without erroring)
type partialMockClient struct {
	data map[string]string
}

func (c *partialMockClient) GetValues(keys []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, k := range keys {
		if v, ok := c.data[k]; ok {
			out[k] = v
		}
		// silently omit missing keys — no error
	}
	return out, nil
}

func (c *partialMockClient) BatchSize() int { return 10 }

func TestLoadKeyNotInMap(t *testing.T) {
	c := &partialMockClient{
		data: map[string]string{"a": "1"},
	}
	var cfg struct {
		A string `ssm:"a"`
		B string `ssm:"b"`
	}
	err := Load(c, &cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load parameter for key 'b'")
}

func TestLoadMalformedTag(t *testing.T) {
	m := NewMockClient()
	var c struct {
		Bad string `ssm:","`
	}
	err := Load(m, &c)
	assert.Error(t, err)
	assert.IsType(t, &TagParseError{}, err)
}

func TestLoadNestedMalformedTag(t *testing.T) {
	m := NewMockClient()
	type Inner struct {
		Bad string `ssm:","`
	}
	var c struct {
		Nested Inner
	}
	err := Load(m, &c)
	assert.Error(t, err)
	assert.IsType(t, &TagParseError{}, err)
}

func TestConvertTypeError_EmptyFieldAndType(t *testing.T) {
	e := &ConvertTypeError{Value: "abc"}
	assert.Equal(t, "failed to convert 'abc'", e.Error())
}

func TestConvertTypeError_WithFieldAndType(t *testing.T) {
	e := &ConvertTypeError{Field: "Name", Type: "int", Value: "abc"}
	assert.Equal(t, "failed to convert 'abc' to int for field Name", e.Error())
}

func TestInvalidTypeError_Ptr(t *testing.T) {
	var x *int
	e := &InvalidTypeError{Type: reflect.TypeOf(x)}
	assert.Equal(t, "invalid type *int", e.Error())
}

func TestBatchIterateEmpty(t *testing.T) {
	var called bool
	err := batchIterateFields(nil, 10, func(f []*field) error {
		called = true
		return nil
	})
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestBatchIterateError(t *testing.T) {
	fields := []*field{{key: "a"}}
	err := batchIterateFields(fields, 10, func(f []*field) error {
		return fmt.Errorf("batch error")
	})
	assert.EqualError(t, err, "batch error")
}

func TestLoadOnlyUnexportedFields(t *testing.T) {
	m := NewMockClient()
	var c struct {
		hidden string //nolint:unused
	}
	err := Load(m, &c)
	assert.NoError(t, err)
}

func TestSSMClientIntegration(t *testing.T) {
	// Test that NewSSMClient produces a working Client that can be passed to Load
	api := &ssmIntegrationMock{
		data: map[string]string{
			"host": "db.example.com",
			"port": "5432",
		},
	}
	c := NewSSMClient(api, false)

	var cfg struct {
		Host string `ssm:"host"`
		Port int    `ssm:"port"`
	}
	err := Load(c, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "db.example.com", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
}

// ssmIntegrationMock is a minimal ssmiface.SSMAPI for integration testing
type ssmIntegrationMock struct {
	ssmiface.SSMAPI
	data map[string]string
}

func (m *ssmIntegrationMock) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	out := &ssm.GetParametersOutput{}
	for _, name := range input.Names {
		k := aws.StringValue(name)
		if v, ok := m.data[k]; ok {
			out.Parameters = append(out.Parameters, &ssm.Parameter{
				Name:  name,
				Value: aws.String(v),
			})
		} else {
			out.InvalidParameters = append(out.InvalidParameters, name)
		}
	}
	return out, nil
}

func TestTagParseWhitespace(t *testing.T) {
	f := reflect.TypeOf(struct {
		Field string `ssm:" key , decrypt , json "`
	}{}).Field(0)
	fld, err := tag(f, nil)
	assert.NoError(t, err)
	assert.Equal(t, "key", fld.key)
	assert.True(t, fld.decrypt)
	assert.True(t, fld.json)
}
