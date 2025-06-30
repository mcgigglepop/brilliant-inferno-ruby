// Package cognito provides a wrapper for AWS Cognito user authentication and management.
package cognito

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitoTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// CognitoClient wraps the AWS Cognito Identity Provider client and configuration.
type CognitoClient struct {
	client      *cognitoidentityprovider.Client // AWS Cognito client
	userPoolID  string                         // Cognito User Pool ID
	clientAppID string                         // Cognito App Client ID
}

// AuthResponse holds authentication tokens returned from Cognito after login.
type AuthResponse struct {
	IdToken      string // JWT ID token
	AccessToken  string // JWT access token
	RefreshToken string // JWT refresh token
}

// NewCognitoClientWithCfg creates a new CognitoClient with the given AWS config, user pool ID, and app client ID.
func NewCognitoClientWithCfg(cfg aws.Config, userPoolID, clientAppID string) (*CognitoClient, error) {
	return &CognitoClient{
		client:      cognitoidentityprovider.NewFromConfig(cfg),
		userPoolID:  userPoolID,
		clientAppID: clientAppID,
	}, nil
}

// RegisterUser registers a new user with the given email and password in Cognito.
func (c *CognitoClient) RegisterUser(ctx context.Context, email, password string) error {
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.clientAppID),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []cognitoTypes.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}

	_, err := c.client.SignUp(ctx, input)
	if err != nil {
		return fmt.Errorf("sign up failed: %w", err)
	}

	return nil
}

// ConfirmUser confirms a user's registration using the provided confirmation code (OTP).
func (c *CognitoClient) ConfirmUser(ctx context.Context, email, confirmationCode string) (*cognitoidentityprovider.ConfirmSignUpOutput, error) {

	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(c.clientAppID),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(confirmationCode),
	}

	result, err := c.client.ConfirmSignUp(ctx, input)
	if err != nil {
		log.Printf("[ERROR] Cognito ConfirmSignUp failed: %v", err)
		return result, err
	}

	return result, nil
}

// Login authenticates a user with email and password, returning authentication tokens.
func (c *CognitoClient) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: "USER_PASSWORD_AUTH",
		ClientId: aws.String(c.clientAppID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	out, err := c.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &AuthResponse{
		IdToken:      aws.ToString(out.AuthenticationResult.IdToken),
		AccessToken:  aws.ToString(out.AuthenticationResult.AccessToken),
		RefreshToken: aws.ToString(out.AuthenticationResult.RefreshToken),
	}, nil
}

// ExtractSubFromToken extracts the user's sub (unique identifier) from a JWT ID token.
func (c *CognitoClient) ExtractSubFromToken(ctx context.Context, idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", err
	}

	sub, ok := payload["sub"].(string)
	if !ok {
		return "", errors.New("sub not found in token")
	}

	return sub, nil
}