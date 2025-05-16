package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	connecthandler "github.com/tikfack/server/internal/infrastructure/connect"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	videoHandler := connecthandler.NewVideoServiceHandler()
	mux := http.NewServeMux()
	pattern, handler := videoHandler.GetHandler()
	mux.Handle(pattern, handler)

	// ミドルウェアチェイン
	loggedHandler := loggingMiddleware(mux)
	handlerWithCORS := cors.AllowAll().Handler(loggedHandler)

	log.Printf("🌐 Connect gRPC server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("サーバー起動に失敗しました: %v", err)
	}
}


func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📥 New request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}