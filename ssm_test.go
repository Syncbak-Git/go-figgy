package figgy

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
)

type mockSSMAPI struct {
	ssmiface.SSMAPI
	params  map[string]string
	invalid []string
	err     error
}

func (m *mockSSMAPI) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := &ssm.GetParametersOutput{}
	for _, name := range input.Names {
		k := aws.StringValue(name)
		if v, ok := m.params[k]; ok {
			out.Parameters = append(out.Parameters, &ssm.Parameter{
				Name:  name,
				Value: aws.String(v),
			})
		}
	}
	for _, inv := range m.invalid {
		out.InvalidParameters = append(out.InvalidParameters, aws.String(inv))
	}
	return out, nil
}

func TestSSMClient_GetValues(t *testing.T) {
	api := &mockSSMAPI{
		params: map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
	}
	c := NewSSMClient(api, false)

	vals, err := c.GetValues([]string{"key1", "key2"})
	assert.NoError(t, err)
	assert.Equal(t, "val1", vals["key1"])
	assert.Equal(t, "val2", vals["key2"])
}

func TestSSMClient_GetValues_APIError(t *testing.T) {
	api := &mockSSMAPI{
		err: fmt.Errorf("connection refused"),
	}
	c := NewSSMClient(api, false)

	_, err := c.GetValues([]string{"key1"})
	assert.EqualError(t, err, "connection refused")
}

func TestSSMClient_GetValues_InvalidParameters(t *testing.T) {
	api := &mockSSMAPI{
		params:  map[string]string{},
		invalid: []string{"bad1", "bad2"},
	}
	c := NewSSMClient(api, true)

	_, err := c.GetValues([]string{"bad1", "bad2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parameters")
	assert.Contains(t, err.Error(), "bad1")
	assert.Contains(t, err.Error(), "bad2")
}

func TestSSMClient_BatchSize(t *testing.T) {
	c := NewSSMClient(&mockSSMAPI{}, false)
	assert.Equal(t, 10, c.BatchSize())
}

func TestSSMClient_DecryptFlag(t *testing.T) {
	var captured *bool
	api := &mockSSMAPI{
		params: map[string]string{"key": "val"},
	}
	// Wrap to capture the WithDecryption value
	wrapper := &captureDecryptSSMAPI{
		mockSSMAPI: api,
		capture:    &captured,
	}

	c := NewSSMClient(wrapper, true)
	_, err := c.GetValues([]string{"key"})
	assert.NoError(t, err)
	assert.NotNil(t, captured)
	assert.True(t, *captured)
}

type captureDecryptSSMAPI struct {
	ssmiface.SSMAPI
	mockSSMAPI *mockSSMAPI
	capture    **bool
}

func (c *captureDecryptSSMAPI) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	*c.capture = input.WithDecryption
	return c.mockSSMAPI.GetParameters(input)
}
