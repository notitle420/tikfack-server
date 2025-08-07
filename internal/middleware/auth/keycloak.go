package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// CallKeycloakPermissionAPI calls UMA endpoint to check if userToken is allowed to access resourceName.
// Doer abstracts HTTP client for testing
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// CallKeycloakPermissionAPIWithClient is a testable version that accepts an HTTP client interface
func CallKeycloakPermissionAPIWithClient(
	ctx context.Context,
	httpClient Doer,
	tokenEndpoint, audience, userToken, resourceName string,
) error {

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:uma-ticket")
	form.Set("audience", audience)
	form.Set("permission", resourceName)
	slog.Info("form", "form", form)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// OK => 許可
		return nil
	}
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("permission denied (403 from Keycloak)")
	}

	return errors.New("unexpected status from keycloak: " + resp.Status)
}

func CallKeycloakPermissionAPI(
	ctx context.Context,
	tokenEndpoint, audience, userToken, resourceName string,
) error {
	return CallKeycloakPermissionAPIWithClient(ctx, http.DefaultClient, tokenEndpoint, audience, userToken, resourceName)
}