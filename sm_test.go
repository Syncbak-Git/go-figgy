package figgy

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/stretchr/testify/assert"
)

type mockSMAPI struct {
	secretsmanageriface.SecretsManagerAPI
	secrets map[string]string
	err     error
}

func (m *mockSMAPI) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	k := aws.StringValue(input.SecretId)
	v, ok := m.secrets[k]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: secret %q not found", k)
	}
	return &secretsmanager.GetSecretValueOutput{
		Name:         aws.String(k),
		SecretString: aws.String(v),
	}, nil
}

func TestSMClient_GetValues(t *testing.T) {
	api := &mockSMAPI{
		secrets: map[string]string{
			"db-host":     "localhost",
			"db-password": "secret123",
		},
	}
	c := NewSecretsManagerClient(api)

	vals, err := c.GetValues([]string{"db-host", "db-password"})
	assert.NoError(t, err)
	assert.Equal(t, "localhost", vals["db-host"])
	assert.Equal(t, "secret123", vals["db-password"])
}

func TestSMClient_GetValues_KeyNotFound(t *testing.T) {
	api := &mockSMAPI{
		secrets: map[string]string{},
	}
	c := NewSecretsManagerClient(api)

	_, err := c.GetValues([]string{"missing"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret errors")
	assert.Contains(t, err.Error(), "missing")
}

func TestSMClient_GetValues_PartialFailure(t *testing.T) {
	api := &mockSMAPI{
		secrets: map[string]string{
			"good-key": "value",
		},
	}
	c := NewSecretsManagerClient(api)

	_, err := c.GetValues([]string{"good-key", "bad-key"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad-key")
}

func TestSMClient_GetValues_AllKeysError(t *testing.T) {
	api := &mockSMAPI{
		err: fmt.Errorf("access denied"),
	}
	c := NewSecretsManagerClient(api)

	_, err := c.GetValues([]string{"key1", "key2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key1")
	assert.Contains(t, err.Error(), "key2")
}

func TestSMClient_BatchSize(t *testing.T) {
	c := NewSecretsManagerClient(&mockSMAPI{})
	assert.Equal(t, 10, c.BatchSize())
}

func TestSMClient_LoadIntegration(t *testing.T) {
	api := &mockSMAPI{
		secrets: map[string]string{
			"app-host": "example.com",
			"app-port": "8080",
			"app-live": "true",
		},
	}
	c := NewSecretsManagerClient(api)

	var cfg struct {
		Host string `ssm:"app-host"`
		Port int    `ssm:"app-port"`
		Live bool   `ssm:"app-live"`
	}
	err := Load(c, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "example.com", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, true, cfg.Live)
}
