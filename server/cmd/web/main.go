package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/cognito"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/config"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/dynamodb"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/handlers"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/helpers"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	// Initialize application
	if err := run(); err != nil {
		log.Fatal(err)
	}

	// Start the HTTP server
	log.Printf("Starting application on port %s", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	// Run the server
	log.Fatal(srv.ListenAndServe())
}

// run initializes the application config, session, AWS clients, and templates
func run() error {
	// Register model for session storage
	gob.Register(map[string]int{})

	// Define command-line flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")

	// Cognito flags
	cognitoUserPoolID := flag.String("cognito-user-pool-id", "", "Cognito user pool ID")
	cognitoClientID := flag.String("cognito-client-id", "", "Cognito app client ID")

	// Parse flags
	flag.Parse()

	// Validate Cognito flags
	if *cognitoUserPoolID == "" || *cognitoClientID == "" {
		fmt.Println("Missing Cognito flags")
		os.Exit(1)
	}

	// Set application config
	app.InProduction = *inProduction
	app.UseCache = *useCache

	// Set up logging
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.InfoLog = infoLog
	app.ErrorLog = errorLog

	// Configure session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// Load AWS config
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("failed to load AWS config:", err)
	}

	// Set Cognito client
	cognitoClient, err := cognito.NewCognitoClientWithCfg(awsCfg, *cognitoUserPoolID, *cognitoClientID)
	if err != nil {
		log.Fatal("failed to create Cognito client:", err)
	}

	app.CognitoClient = cognitoClient

	// Set DynamoDB clients
	app.Dynamo = &config.DynamoService{
		// set up dynamodb clients here
	}

	// Create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return err
	}
	app.TemplateCache = tc

	// Initialize handlers
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return nil
}