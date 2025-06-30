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

type CognitoClient struct {
	client      *cognitoidentityprovider.Client
	userPoolID  string
	clientAppID string
}

type AuthResponse struct {
	IdToken      string
	AccessToken  string
	RefreshToken string
}

func NewCognitoClientWithCfg(cfg aws.Config, userPoolID, clientAppID string) (*CognitoClient, error) {
	return &CognitoClient{
		client:      cognitoidentityprovider.NewFromConfig(cfg),
		userPoolID:  userPoolID,
		clientAppID: clientAppID,
	}, nil
}

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