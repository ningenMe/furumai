// Package dynamodb is furumai's DynamoDB Stimulus/Observation adapter.
package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Stimulus writes items to DynamoDB tables. It is used from both given and
// when steps, since Stimulus adapters are shared between them, and also
// serves as the Observation side via GetItem/Scan.
type Stimulus struct {
	Client *dynamodb.Client
}

// NewStimulus loads the default AWS config (credentials, region, ...) and
// returns a Stimulus. optFns customize the underlying DynamoDB client,
// e.g. to point at a non-AWS DynamoDB-compatible endpoint:
//
//	dynamodb.NewStimulus(ctx, func(o *awsdynamodb.Options) {
//		o.BaseEndpoint = aws.String("http://localhost:4566")
//	})
func NewStimulus(ctx context.Context, optFns ...func(*dynamodb.Options)) (*Stimulus, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &Stimulus{Client: dynamodb.NewFromConfig(cfg, optFns...)}, nil
}

// PutItem writes item to table. Matchers aren't accepted here (item is
// marshaled to DynamoDB's wire format), only literal values.
func (s *Stimulus) PutItem(ctx context.Context, table string, item map[string]any) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = s.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      av,
	})
	return err
}

// DeleteItem removes the item identified by key (its primary key
// attributes) from table.
func (s *Stimulus) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	av, err := attributevalue.MarshalMap(key)
	if err != nil {
		return err
	}
	_, err = s.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key:       av,
	})
	return err
}

// Item is the full-state Observation for a single DynamoDB item: attribute
// name to value. Matchers (Any, Regex, ...) can be substituted for any
// attribute since map values are any-typed.
type Item map[string]any

// GetItem returns the full-state Observation for the item identified by
// key, or (nil, nil) if no item exists. To assert non-existence, compare
// the result against a plain nil check rather than through
// furumai.ThenEqual (see adapter/s3.GetObject for the same pattern).
func (s *Stimulus) GetItem(ctx context.Context, table string, key map[string]any) (Item, error) {
	keyAV, err := attributevalue.MarshalMap(key)
	if err != nil {
		return nil, err
	}

	out, err := s.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       keyAV,
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, nil
	}

	var item Item
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, err
	}
	return item, nil
}

// Scan returns every item in table as the full-state Observation. Compare
// the result against an expected []Item with furumai.Diff/furumai.ThenEqual
// (wrap the expectation in furumai.AnyOrder, since Scan order isn't
// guaranteed).
func (s *Stimulus) Scan(ctx context.Context, table string) ([]Item, error) {
	out, err := s.Client.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String(table)})
	if err != nil {
		return nil, err
	}

	items := make([]Item, 0, len(out.Items))
	for _, raw := range out.Items {
		var item Item
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
