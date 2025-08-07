package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/coreos/go-oidc"
	gocloak "github.com/mviniciusgc/gocloak/v13"
	"github.com/tikfack/server/internal/middleware/auth/mock"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// OIDCInterceptor validates Authorization header on every RPC and stores sub in context.
// エラータイプの定義
var (
	ErrInvalidAuthHeader = errors.New("invalid auth header")
	ErrTokenNotActive    = errors.New("token is not active")
	ErrNoPermission      = errors.New("insufficient permissions")
	ErrNoTokenInContext  = errors.New("no token in context")
	ErrNoResourceMapping = errors.New("no resource mapping for method")
	ErrNoRPTReturned     = errors.New("no RPT (permission ticket) returned")
)

// extractBearerToken extracts Bearer token from Authorization header
func extractBearerToken(req connect.AnyRequest) (string, error) {
	header := req.Header().Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}
	return parts[1], nil
}

// ResourceMapper maps method names to resource names
type ResourceMapper interface {
	GetResource(methodName string) (string, error)
}

// DefaultResourceMapper implements ResourceMapper with predefined mappings
type DefaultResourceMapper struct {
	mappings map[string]string
}

// NewDefaultResourceMapper creates a new DefaultResourceMapper with default mappings
func NewDefaultResourceMapper() *DefaultResourceMapper {
	return &DefaultResourceMapper{
		mappings: map[string]string{
			"GetVideosByKeyword": "resource-get-videos-by-keyword",
			"GetVideosByDate":    "resource-get-videos-by-date",
		},
	}
}

// GetResource returns the resource name for a given method name
func (m *DefaultResourceMapper) GetResource(methodName string) (string, error) {
	// Extract method name from full procedure path
	parts := strings.Split(methodName, "/")
	if len(parts) > 0 {
		methodName = parts[len(parts)-1]
	}
	
	resource, exists := m.mappings[methodName]
	if !exists {
		return "", ErrNoResourceMapping
	}
	return resource, nil
}

// AddMapping adds a new method-to-resource mapping
func (m *DefaultResourceMapper) AddMapping(methodName, resourceName string) {
	m.mappings[methodName] = resourceName
}

func OIDCInterceptor(verifier *oidc.IDTokenVerifier) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			rawToken, err := extractBearerToken(req)
			if err != nil {
				slog.Error("failed to extract bearer token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			idt, err := verifier.Verify(ctx, rawToken)
			// Token verification successful
			if err != nil {
				slog.Error("failed to verify idt", "err", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
				slog.Error("failed to extract claims", "error", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			ctx = context.WithValue(ctx, ctxkeys.TokenKey, rawToken)
			ctx = context.WithValue(ctx, ctxkeys.SubKey, claims.Sub)
			return next(ctx, req)
		}
	})
}

func PermissionInterceptor(
	client *gocloak.GoCloak,
	realm string,
	clientID string,
	checkPermission func(
		ctx context.Context,
		client *gocloak.GoCloak,
		userToken string,
		resourceName string,
		realm string,
		clientID string,
	) error,
) connect.Interceptor {
	return PermissionInterceptorWithMapper(client, realm, clientID, checkPermission, NewDefaultResourceMapper())
}

// PermissionInterceptorWithMapper allows custom resource mapping
func PermissionInterceptorWithMapper(
	client *gocloak.GoCloak,
	realm string,
	clientID string,
	checkPermission func(
		ctx context.Context,
		client *gocloak.GoCloak,
		userToken string,
		resourceName string,
		realm string,
		clientID string,
	) error,
	mapper ResourceMapper,
) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {

			// 1) Token取得
			tokenVal := ctx.Value(ctxkeys.TokenKey)
			if tokenVal == nil {
				slog.Error("no token found in context")
				return nil, connect.NewError(connect.CodeUnauthenticated, ErrNoTokenInContext)
			}
			userToken, _ := tokenVal.(string)
			
			// 2) メソッド名 → resourceNameを決定
			methodFullName := req.Spec().Procedure
			var resourceName string
			switch {
			case strings.HasSuffix(methodFullName, "GetVideosByKeyword"):
				resourceName = "resource-get-videos-by-keyword"
			case strings.HasSuffix(methodFullName, "GetVideosByDate"):
				resourceName = "resource-get-videos-by-date"
			default:
				return nil, connect.NewError(connect.CodePermissionDenied,
					errors.New("no resource mapping for "+methodFullName))
			}

			// 3) Keycloak へ問い合わせ
			if err := checkPermission(ctx, client, userToken, resourceName, realm, clientID); err != nil {
				slog.Warn("permission denied for user", "resource", resourceName, "error", err)
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}
			return next(ctx, req)
		}
	})
}

// IntrospectionInterceptorWithInterfaces is a testable version that accepts interfaces
func IntrospectionInterceptorWithInterfaces(
	verifier mock.IDTokenVerifierInterface,
	client mock.GocloakClientInterface,
	realm,
	clientID,
	clientSecret string,
) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// 1) Authorization ヘッダ取得
			token, err := extractBearerToken(req)
			if err != nil {
				slog.Error("failed to extract bearer token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// 2) go-oidc で署名検証 & sub 取得
			idt, err := verifier.Verify(ctx, token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
				slog.Error("failed to extract claims", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// 3) IntrospectToken（＝Token Introspection エンドポイント呼び出し）
			result, err := client.RetrospectToken(ctx, token, clientID, clientSecret, realm)
			if err != nil {
				slog.Error("failed to introspect token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			if result.Active == nil || !*result.Active {
				slog.Warn("token is not active")
				return nil, connect.NewError(connect.CodeUnauthenticated, ErrTokenNotActive)
			}

			ctx = context.WithValue(ctx, ctxkeys.TokenKey, token)
			ctx = context.WithValue(ctx, ctxkeys.SubKey, claims.Sub)

			return next(ctx, req)
		}
	})
}

func IntrospectionInterceptor(
	verifier *oidc.IDTokenVerifier,
	client *gocloak.GoCloak,
	realm,
	clientID,
	clientSecret string,
) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// 1) Authorization ヘッダ取得
			token, err := extractBearerToken(req)
			if err != nil {
				slog.Error("failed to extract bearer token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// 2) go-oidc で署名検証 & sub 取得
			idt, err := verifier.Verify(ctx, token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
				slog.Error("failed to extract claims", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// 3) IntrospectToken（＝Token Introspection エンドポイント呼び出し）
			result, err := client.RetrospectToken(ctx, token, clientID, clientSecret, realm)
			if err != nil {
				slog.Error("failed to introspect token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			if result.Active == nil || !*result.Active {
				slog.Warn("token is not active")
				return nil, connect.NewError(connect.CodeUnauthenticated, ErrTokenNotActive)
			}

			ctx = context.WithValue(ctx, ctxkeys.TokenKey, token)
			ctx = context.WithValue(ctx, ctxkeys.SubKey, claims.Sub)

			return next(ctx, req)
		}
	})
}

// CheckPermissionFuncWithInterface is a testable version that accepts interfaces
func CheckPermissionFuncWithInterface(ctx context.Context, client mock.GocloakClientInterface, userToken, resourceName, realm, clientID string) error {
	options := gocloak.RequestingPartyTokenOptions{
		Audience:    gocloak.StringP(clientID),
		Permissions: &[]string{resourceName},
	}

	// userToken (access_token) を持って UMA Permission API を叩く
	rpt, err := client.GetRequestingPartyToken(ctx, userToken, realm, options)
	if err != nil {
		slog.Error("failed to get requesting party token", "error", err)
		return err
	}
	if rpt.AccessToken == "" {
		slog.Error("no RPT returned from Keycloak")
		return ErrNoRPTReturned
	}
	// 200 が返ってきて rpt.AccessToken がセットされていれば許可
	return nil
}

func CheckPermissionFunc(ctx context.Context, client *gocloak.GoCloak, userToken, resourceName, realm, clientID string) error { // Keycloak のベース URL
	wrapper := mock.NewGocloakClientWrapper(client)
	return CheckPermissionFuncWithInterface(ctx, wrapper, userToken, resourceName, realm, clientID)
}
