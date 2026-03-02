package figgy

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

const maxSSMParameters = 10

type ssmClient struct {
	api     ssmiface.SSMAPI
	decrypt bool
}

// NewSSMClient wraps an SSM API client as a figgy Client.
// When decrypt is true, all parameters are fetched with WithDecryption enabled.
func NewSSMClient(api ssmiface.SSMAPI, decrypt bool) Client {
	return &ssmClient{api: api, decrypt: decrypt}
}

func (c *ssmClient) GetValues(keys []string) (map[string]string, error) {
	names := make([]*string, len(keys))
	for i, k := range keys {
		names[i] = aws.String(k)
	}
	res, err := c.api.GetParameters(&ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(c.decrypt),
	})
	if err != nil {
		return nil, err
	}
	if len(res.InvalidParameters) != 0 {
		return nil, fmt.Errorf("invalid parameters: %s",
			strings.Join(aws.StringValueSlice(res.InvalidParameters), ", "),
		)
	}
	out := make(map[string]string, len(res.Parameters))
	for _, p := range res.Parameters {
		out[aws.StringValue(p.Name)] = aws.StringValue(p.Value)
	}
	return out, nil
}

func (c *ssmClient) BatchSize() int { return maxSSMParameters }
