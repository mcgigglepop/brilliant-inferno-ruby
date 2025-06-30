// Package dynamodb provides a DynamoDB client wrapper for application data access.
package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	sdkdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/models"
)

// DDBClient wraps the AWS DynamoDB client and table name for application use.
type DDBClient struct {
	db        *sdkdynamodb.Client // DynamoDB client
	tableName string             // Table name used for operations
}

// NewAppClient creates a new DDBClient with the given AWS config and table name.
func NewAppClient(cfg aws.Config, tableName string) *DDBClient {
	return &DDBClient{
		db:        sdkdynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}
}