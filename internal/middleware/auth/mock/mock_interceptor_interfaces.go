package mock

import (
	"context"

	"github.com/coreos/go-oidc"
	gocloak "github.com/mviniciusgc/gocloak/v13"
)

//go:generate mockgen -source=interfaces.go -destination=mock_interfaces.go -package=mock

// IDTokenInterface wraps oidc.IDToken for testing
type IDTokenInterface interface {
	Claims(v interface{}) error
}

// IDTokenVerifierInterface wraps oidc.IDTokenVerifier for mocking
type IDTokenVerifierInterface interface {
	Verify(ctx context.Context, rawIDToken string) (IDTokenInterface, error)
}

// GocloakClientInterface wraps gocloak.GoCloak methods for mocking
type GocloakClientInterface interface {
	RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error)
	GetRequestingPartyToken(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*gocloak.JWT, error)
}

// IDTokenVerifierWrapper wraps the actual oidc.IDTokenVerifier
type IDTokenVerifierWrapper struct {
	verifier *oidc.IDTokenVerifier
}

func NewIDTokenVerifierWrapper(verifier *oidc.IDTokenVerifier) *IDTokenVerifierWrapper {
	return &IDTokenVerifierWrapper{verifier: verifier}
}

func (w *IDTokenVerifierWrapper) Verify(ctx context.Context, rawIDToken string) (IDTokenInterface, error) {
	return w.verifier.Verify(ctx, rawIDToken)
}

// GocloakClientWrapper wraps the actual gocloak.GoCloak
type GocloakClientWrapper struct {
	client *gocloak.GoCloak
}

func NewGocloakClientWrapper(client *gocloak.GoCloak) *GocloakClientWrapper {
	return &GocloakClientWrapper{client: client}
}

func (w *GocloakClientWrapper) RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return w.client.RetrospectToken(ctx, accessToken, clientID, clientSecret, realm)
}

func (w *GocloakClientWrapper) GetRequestingPartyToken(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*gocloak.JWT, error) {
	return w.client.GetRequestingPartyToken(ctx, token, realm, options)
}