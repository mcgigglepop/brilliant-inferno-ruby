// Package config provides application configuration structures and dependencies.
package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/cognito"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/dynamodb"
)

// DynamoService is a placeholder for DynamoDB-related services or methods.
type DynamoService struct {
}

// AppConfig holds the application configuration and shared dependencies.
type AppConfig struct {
	UseCache      bool                          // Whether to use the template cache
	TemplateCache map[string]*template.Template // Cached templates
	InfoLog       *log.Logger                  // Logger for informational messages
	ErrorLog      *log.Logger                  // Logger for error messages
	InProduction  bool                         // True if running in production
	Session       *scs.SessionManager          // Session manager
	CognitoClient *cognito.CognitoClient       // AWS Cognito client for authentication
	Dynamo        *DynamoService               // DynamoDB service wrapper
}