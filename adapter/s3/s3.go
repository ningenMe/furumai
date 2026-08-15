// Package s3 is furumai's Object Storage (S3) Stimulus/Observation
// adapter.
package s3

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Stimulus writes objects to Bucket. It is used from both given and when
// steps, since Stimulus adapters are shared between them, and also serves
// as the Observation side via GetObject.
type Stimulus struct {
	Client *s3.Client
	Bucket string
}

// NewStimulus loads the default AWS config (credentials, region, ...) and
// returns a Stimulus for bucket. optFns customize the underlying S3
// client, e.g. to point at a non-AWS S3-compatible endpoint:
//
//	s3.NewStimulus(ctx, "my-bucket", func(o *awss3.Options) {
//		o.BaseEndpoint = aws.String("http://localhost:4566")
//		o.UsePathStyle = true
//	})
func NewStimulus(ctx context.Context, bucket string, optFns ...func(*s3.Options)) (*Stimulus, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &Stimulus{Client: s3.NewFromConfig(cfg, optFns...), Bucket: bucket}, nil
}

type putOptions struct {
	metadata map[string]string
}

// PutOption customizes PutObject.
type PutOption func(*putOptions)

// WithMetadata attaches a user metadata key/value to the object.
func WithMetadata(key, value string) PutOption {
	return func(o *putOptions) {
		if o.metadata == nil {
			o.metadata = make(map[string]string)
		}
		o.metadata[key] = value
	}
}

// PutObject writes content to key.
func (s *Stimulus) PutObject(ctx context.Context, key string, content []byte, opts ...PutOption) error {
	var o putOptions
	for _, opt := range opts {
		opt(&o)
	}

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(key),
		Body:     bytes.NewReader(content),
		Metadata: o.metadata,
	})
	return err
}

// Object is the full-state Observation for a single object. Content and
// Metadata are typed any so a furumai.Matcher (Any, Regex, ...) can be
// substituted for either.
type Object struct {
	Key      string
	Content  any
	Metadata any
}

// GetObject returns the full-state Observation for key, or (nil, nil) if
// no object exists at key. To assert non-existence, compare the result
// against a plain nil check rather than through furumai.ThenEqual, e.g.:
//
//	got, err := s.GetObject(ctx, "missing-key")
//	if got != nil { t.Fatalf("expected object to not exist") }
func (s *Stimulus) GetObject(ctx context.Context, key string) (*Object, error) {
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, nil
		}
		return nil, err
	}
	defer out.Body.Close()

	content, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}

	return &Object{
		Key:      key,
		Content:  string(content),
		Metadata: out.Metadata,
	}, nil
}
