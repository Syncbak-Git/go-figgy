package figgy

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type smClient struct {
	api secretsmanageriface.SecretsManagerAPI
}

// NewSecretsManagerClient wraps a Secrets Manager API client as a figgy Client.
func NewSecretsManagerClient(api secretsmanageriface.SecretsManagerAPI) Client {
	return &smClient{api: api}
}

func (c *smClient) GetValues(keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	var errs []string
	for _, k := range keys {
		res, err := c.api.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(k),
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", k, err))
			continue
		}
		out[aws.StringValue(res.Name)] = aws.StringValue(res.SecretString)
	}
	if len(errs) != 0 {
		return nil, fmt.Errorf("secret errors: %s", strings.Join(errs, ", "))
	}
	return out, nil
}

// BatchSize returns the number of keys to fetch per batch. Since Secrets Manager
// GetSecretValue is a single-key API, batching is handled internally in GetValues,
// so we use a reasonable batch size to limit the number of keys per call to load().
func (c *smClient) BatchSize() int { return 10 }
