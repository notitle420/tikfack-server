package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/coreos/go-oidc"
	gocloak "github.com/mviniciusgc/gocloak/v13"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)

// OIDCInterceptor validates Authorization header on every RPC and stores sub in context.
func OIDCInterceptor(verifier *oidc.IDTokenVerifier) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			header := req.Header().Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				slog.Error("invalid auth header", "parts", parts)
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth header"))
			}
			var rawToken = parts[1]
			idt, err := verifier.Verify(ctx, rawToken)
			slog.Debug("idt", "idt", idt)
			if err != nil {
				slog.Error("failed to verify idt", "err", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
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
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {

			// 1) Token取得
			tokenVal := ctx.Value(ctxkeys.TokenKey)
			if tokenVal == nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no token in context"))
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
				slog.Info("Not Permission")
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}
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
			header := req.Header().Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth header"))
			}
			token := parts[1]

			// 2) go-oidc で署名検証 & sub 取得
			idt, err := verifier.Verify(ctx, token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// 2) IntrospectToken（＝Token Introspection エンドポイント呼び出し）
			result, err := client.RetrospectToken(ctx, token, clientID, clientSecret, realm)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			if result.Active == nil || !*result.Active {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token is not active"))
			}

			ctx = context.WithValue(ctx, ctxkeys.TokenKey, token)
			ctx = context.WithValue(ctx, ctxkeys.SubKey, claims.Sub)

			return next(ctx, req)
		}
	})
}

func CheckPermissionFunc(ctx context.Context, client *gocloak.GoCloak, userToken, resourceName, realm, clientID string) error { // Keycloak のベース URL
	options := gocloak.RequestingPartyTokenOptions{
		Audience:    gocloak.StringP(clientID),
		Permissions: &[]string{resourceName},
	}

	// userToken (access_token) を持って UMA Permission API を叩く
	rpt, err := client.GetRequestingPartyToken(ctx, userToken, realm, options)
	if err != nil {
		slog.Error("err", "err", err)
		return err
	}
	if rpt.AccessToken == "" {
		return errors.New("no RPT (permission ticket) returned")
	}
	// 200 が返ってきて rpt.AccessToken がセットされていれば許可
	return nil
}
