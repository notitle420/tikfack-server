package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/tikfack/server/internal/infrastructure/auth"
	connecthandler "github.com/tikfack/server/internal/infrastructure/connect"
)

func main() {
	// .env の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	ctx := context.Background()

	verifier, err := auth.NewVerifier(ctx, "http://localhost:8080/realms/myrealm", "backend-service")
	if err != nil {
		log.Fatalf("OIDC verifier init failed: %v", err)
	}

	oidcInterceptor := auth.OIDCInterceptor(verifier)

	handler := connecthandler.NewVideoServiceHandler(connect.WithInterceptors(oidcInterceptor))

	mux := http.NewServeMux()
	pattern, h := handler.GetHandler()
	mux.Handle(pattern, h)

	handlerWithCORS := cors.AllowAll().Handler(mux)
	log.Printf("Connect gRPC server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
