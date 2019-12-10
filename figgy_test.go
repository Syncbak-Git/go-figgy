package figgy

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type MockSSMClient struct {
	ssmiface.SSMAPI
	Data map[string]*ssm.GetParameterOutput
}

func (c MockSSMClient) GetParameter(i *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	//TODO: Lookup key and mimic more closely how the aws sdk works, no key causes a panic
	return c.Data[*i.Name], nil
}

func (c MockSSMClient) GetParameters(i *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	var out = new(ssm.GetParametersOutput)
	if len(i.Names) > maxParameters {
		return nil, fmt.Errorf("max parameters exceeded: received %d, max %d", len(i.Names), maxParameters)
	}
	for _, n := range i.Names {
		p, ok := c.Data[aws.StringValue(n)]
		if !ok {
			out.InvalidParameters = append(out.InvalidParameters, n)
		} else {
			out.Parameters = append(out.Parameters, p.Parameter)
		}
	}
	return out, nil
}

func NewMockSSMClient() *MockSSMClient {
	m := &MockSSMClient{}
	m.Data = map[string]*ssm.GetParameterOutput{
		"bool": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("bool"),
				Type:  aws.String("string"),
				Value: aws.String("true"),
			},
		},
		"int": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("int"),
				Type:  aws.String("string"),
				Value: aws.String("2"),
			},
		},
		"int8": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("int8"),
				Type:  aws.String("string"),
				Value: aws.String("3"),
			},
		},
		"int16": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("int16"),
				Type:  aws.String("string"),
				Value: aws.String("4"),
			},
		},
		"int32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("int32"),
				Type:  aws.String("string"),
				Value: aws.String("5"),
			},
		},
		"int64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("int64"),
				Type:  aws.String("string"),
				Value: aws.String("6"),
			},
		},
		"uint": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uint"),
				Type:  aws.String("string"),
				Value: aws.String("7"),
			},
		},
		"uint8": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uint8"),
				Type:  aws.String("string"),
				Value: aws.String("8"),
			},
		},
		"uint16": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uint16"),
				Type:  aws.String("string"),
				Value: aws.String("9"),
			},
		},
		"uint32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uint32"),
				Type:  aws.String("string"),
				Value: aws.String("10"),
			},
		},
		"uint64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uint64"),
				Type:  aws.String("string"),
				Value: aws.String("11"),
			},
		},
		"uintptr": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("uintptr"),
				Type:  aws.String("string"),
				Value: aws.String("12"),
			},
		},
		"float32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("float32"),
				Type:  aws.String("string"),
				Value: aws.String("12.1"),
			},
		},
		"float64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("float64"),
				Type:  aws.String("string"),
				Value: aws.String("12.2"),
			},
		},
		"duration": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("duration"),
				Type:  aws.String("string"),
				Value: aws.String("3600000000000"),
			},
		},
		"durationstring": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("durationstring"),
				Type:  aws.String("string"),
				Value: aws.String("3600s"),
			},
		},
		"pbool": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pbool"),
				Type:  aws.String("string"),
				Value: aws.String("true"),
			},
		},
		"pint": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pint"),
				Type:  aws.String("string"),
				Value: aws.String("13"),
			},
		},
		"pint8": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pint8"),
				Type:  aws.String("string"),
				Value: aws.String("14"),
			},
		},
		"pint16": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pint16"),
				Type:  aws.String("string"),
				Value: aws.String("15"),
			},
		},
		"pint32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pint32"),
				Type:  aws.String("string"),
				Value: aws.String("16"),
			},
		},
		"pint64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pint64"),
				Type:  aws.String("string"),
				Value: aws.String("17"),
			},
		},
		"puint": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puint"),
				Type:  aws.String("string"),
				Value: aws.String("18"),
			},
		},
		"puint8": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puint8"),
				Type:  aws.String("string"),
				Value: aws.String("19"),
			},
		},
		"puint16": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puint16"),
				Type:  aws.String("string"),
				Value: aws.String("20"),
			},
		},
		"puint32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puint32"),
				Type:  aws.String("string"),
				Value: aws.String("21"),
			},
		},
		"puint64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puint64"),
				Type:  aws.String("string"),
				Value: aws.String("22"),
			},
		},
		"puintptr": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("puintptr"),
				Type:  aws.String("string"),
				Value: aws.String("23"),
			},
		},
		"pfloat32": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pfloat32"),
				Type:  aws.String("string"),
				Value: aws.String("23.1"),
			},
		},
		"pfloat64": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pfloat64"),
				Type:  aws.String("string"),
				Value: aws.String("23.2"),
			},
		},
		"string": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("string"),
				Type:  aws.String("string"),
				Value: aws.String("this is a string"),
			},
		},
		"pstring": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pstring"),
				Type:  aws.String("string"),
				Value: aws.String("this is a ptr to a string"),
			},
		},
		"sliceint": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("sliceint"),
				Type:  aws.String("string"),
				Value: aws.String("1,2,3,4,5"),
			},
		},
		"pduration": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pduration"),
				Type:  aws.String("string"),
				Value: aws.String("3600000000000"),
			},
		},
		"pdurationString": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("pdurationstring"),
				Type:  aws.String("string"),
				Value: aws.String("3600s"),
			},
		},
		"simplejson": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("simplejson"),
				Type:  aws.String("string"),
				Value: aws.String(`{"F1": 1, "F2": "2"}`),
			},
		},
		"simplejsonarray": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("simplejsonarray"),
				Type:  aws.String("string"),
				Value: aws.String(`[{"F1": 1, "F2": "2"}]`),
			},
		},
		"badjson": {
			Parameter: &ssm.Parameter{
				Name:  aws.String("badjson"),
				Type:  aws.String("string"),
				Value: aws.String("invalid"),
			},
		},
	}
	return m
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

	//UintptrStr uintptr

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
	/*
		SliceN  []Nested
		SlicePN []*Nested
		PSliceN *[]Nested

		Interface  interface{}
		PInterface *interface{}
	*/
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

	m := NewMockSSMClient()
	for n, tc := range tests {
		_, err := Load(m, tc.in)
		assert.EqualErrorf(t, err, tc.want.Error(), "unexpected error while executing test %s", n)
	}
}

func TestTypeConvert(t *testing.T) {
	ex := NewTypes()
	_, err := Load(NewMockSSMClient(), ex)
	if err != nil {
		t.Fatal(err)
	}
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
		_, err := Load(NewMockSSMClient(), tc.in)
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
		_, err := Load(NewMockSSMClient(), tc.in)
		assert.EqualError(t, err, tc.want.Error(), "test '%s' failed", n)
	}
}

func TestInvalidParams(t *testing.T) {
	var c struct {
		Invalid string `ssm:"/no/such/param"`
	}
	_, err := Load(NewMockSSMClient(), &c)
	assert.Error(t, err)
}

func TestMixedPlainAndDecryptParams(t *testing.T) {
	var c struct {
		Plain1   string `ssm:"string"`
		Plain2   bool   `ssm:"bool"`
		Decrypt1 int    `ssm:"int,decrypt"`
		Decrypt2 int32  `ssm:"int32,decrypt"`
	}
	_, err := Load(NewMockSSMClient(), &c)
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
	_, err := Load(NewMockSSMClient(), &j)
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
	_, err := Load(NewMockSSMClient(), &j)
	assert.Error(t, err)
}

func TestJSONWithUnmarshallerError(t *testing.T) {
	var j struct {
		Test str `ssm:"string,json"`
	}
	_, err := Load(NewMockSSMClient(), &j)
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

func TestPartition(t *testing.T) {
	var tests = []struct {
		in   []bool
		lenp int
		lend int
	}{
		{nil, 0, 0},
		{[]bool{}, 0, 0},
		{[]bool{false}, 1, 0},
		{[]bool{true}, 0, 1},
		{[]bool{false, true}, 1, 1},
		{[]bool{true, false}, 1, 1},
		{[]bool{false, false}, 2, 0},
		{[]bool{true, true}, 0, 2},
		{[]bool{true, false, true}, 1, 2},
		{[]bool{false, true, false}, 2, 1},
		{[]bool{false, false, true}, 2, 1},
		{[]bool{true, false, false}, 2, 1},
		{[]bool{false, false, false}, 3, 0},
		{[]bool{true, true, true}, 0, 3},
	}
	for _, x := range tests {
		f := makePartitionFields(x.in)
		plain, decrypt := partitionFields(f, func(x *field) bool {
			return x.decrypt
		})
		assert.Len(t, plain, x.lenp)
		assert.Len(t, decrypt, x.lend)
		for i := range plain {
			assert.Equal(t, false, plain[i].decrypt)
		}
		for i := range decrypt {
			assert.Equal(t, true, decrypt[i].decrypt)
		}
	}
}

func makePartitionFields(x []bool) []*field {
	if x == nil {
		return nil
	}
	f := make([]*field, len(x))
	for i := range x {
		f[i] = &field{decrypt: x[i]}
	}
	return f
}
