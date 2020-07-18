package storage

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// PMSKMS implements storage.Provider and storage.SyncableProvider and uses
// AWS Parameter Store and AWS KMS to store secrets.
type PMSKMS struct {
	svc   *ssm.SSM
	keyID string
}

func NewPMSKMS(sess *session.Session, keyID string) *PMSKMS {
	return &PMSKMS{
		keyID: keyID,
		svc:   ssm.New(sess),
	}
}

func (s *PMSKMS) Get(ctx context.Context, ks []string) (out []Result, err error) {
	names := make([]*string, 0, len(ks))

	for _, k := range ks {
		names = append(names, aws.String(k))
	}

	o, err := s.svc.GetParametersWithContext(ctx, &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	for _, v := range o.Parameters {
		out = append(out, Result{
			Value:   *v.Value,
			Version: int(*v.Version),
		})
	}

	return
}

func (s *PMSKMS) Set(ctx context.Context, k, v string) error {
	_, err := s.svc.PutParameterWithContext(ctx, &ssm.PutParameterInput{
		KeyId:     aws.String(s.keyID),
		Name:      aws.String(k),
		Overwrite: aws.Bool(true),
		Tier:      aws.String("Standard"),
		Type:      aws.String("SecureString"),
		Value:     aws.String(v),
	})
	return err
}

func (s *PMSKMS) LatestVersion(ctx context.Context, key string) (int, error) {
	res, err := s.svc.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		return 0, err
	}

	return int(*res.Parameter.Version), nil
}

func (s *PMSKMS) GetByPrefix(ctx context.Context, prefix string) (keys []string, err error) {
	return keys, s.describeParamsPaginator(ctx, prefix, nil, &keys)
}

func (s *PMSKMS) describeParamsPaginator(ctx context.Context, prefix string, token *string, rs *[]string) error {
	res, err := s.svc.DescribeParametersWithContext(ctx, &ssm.DescribeParametersInput{
		NextToken: token,
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("BeginsWith"),
				Values: []*string{
					aws.String(prefix),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, v := range res.Parameters {
		*rs = append(*rs, *v.Name)
	}

	if res.NextToken == nil || len(res.Parameters) == 0 {
		return nil
	}

	return s.describeParamsPaginator(ctx, prefix, res.NextToken, rs)
}
