package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	"go.uber.org/mock/gomock"

	// protoから生成したパッケージ
	"context"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	pb "github.com/tikfack/server/gen/video"
	videopbconnect "github.com/tikfack/server/gen/video/videoconnect"
	mockpb "github.com/tikfack/server/internal/infrastructure/connect/mock"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		// .envファイルがなくてもエラーではない（本番環境では環境変数で設定する場合がある）
		slog.Info("環境変数を.envから読み込めませんでした", "error", err)
	}

	// ログレベルの設定
	logLevel := os.Getenv("LOG_LEVEL")
	setupLogger(logLevel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

    ctrl := gomock.NewController(nil)
	mockHandler := mockpb.NewMockVideoServiceHandler(ctrl)

	// モックの実装
	mockHandler.EXPECT().
		GetVideosByDate(gomock.Any(), gomock.Any()).
        DoAndReturn(func(ctx context.Context, req *connect.Request[pb.GetVideosByDateRequest]) (*connect.Response[pb.GetVideosByDateResponse], error) {
            videos := []*pb.Video{
                {
                    DmmId:        "DUMMY001",
                    Title:        "サンプル動画タイトル1",
                    DirectUrl:    "https://example.com/videos/1",
                    Url:          "https://dmm.com/video/1",
                    SampleUrl:    "https://example.com/sample/1.mp4",
                    ThumbnailUrl: "https://example.com/thumbs/1.jpg",
                    CreatedAt:    "2024-05-19T00:00:00Z",
                    Price:        2980,
                    LikesCount:   123,
    
                    Actresses: []*pb.Actress{
                        {Id: "A001", Name: "佐藤花子"},
                        {Id: "A002", Name: "鈴木美咲"},
                    },
                    Genres: []*pb.Genre{
                        {Id: "G001", Name: "ドラマ"},
                        {Id: "G002", Name: "ロマンス"},
                    },
                    Makers: []*pb.Maker{
                        {Id: "M001", Name: "メーカーテスト"},
                    },
                    Series: []*pb.Series{
                        {Id: "S001", Name: "シリーズサンプル"},
                    },
                    Directors: []*pb.Director{
                        {Id: "D001", Name: "山田太郎"},
                    },
                    Review: &pb.Review{
                        Count:     100,
                        Average:  4.5,
                    },
                },
                {
                    DmmId:        "DUMMY002",
                    Title:        "サンプル動画タイトル2",
                    DirectUrl:    "https://example.com/videos/2",
                    Url:          "https://dmm.com/video/2",
                    SampleUrl:    "https://example.com/sample/2.mp4",
                    ThumbnailUrl: "https://example.com/thumbs/2.jpg",
                    CreatedAt:    "2024-05-18T00:00:00Z",
                    Price:        1980,
                    LikesCount:   87,
    
                    Actresses: []*pb.Actress{
                        {Id: "A003", Name: "高橋彩"},
                    },
                    Genres: []*pb.Genre{
                        {Id: "G003", Name: "アクション"},
                    },
                    Makers: []*pb.Maker{
                        {Id: "M002", Name: "テストメーカー2"},
                    },
                    Series: []*pb.Series{
                        {Id: "S002", Name: "シリーズ2"},
                    },
                    Directors: []*pb.Director{
                        {Id: "D002", Name: "佐々木健"},
                    },
                    Review: &pb.Review{
                        Count:     100,
                        Average:  3.0,
                    },
                },
            }
            resp := &pb.GetVideosByDateResponse{Videos: videos}
            return connect.NewResponse(resp), nil
        }).
        AnyTimes()

    pattern, handler := videopbconnect.NewVideoServiceHandler(mockHandler)
	mux := http.NewServeMux()
	mux.Handle(pattern, handler)

	// ミドルウェアチェイン
	loggedHandler := loggingMiddleware(mux)
	handlerWithCORS := cors.AllowAll().Handler(loggedHandler)

	slog.Info("サーバーを起動しています", "port", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		slog.Error("サーバー起動に失敗しました", "error", err)
		os.Exit(1)
	}
}

// setupLogger configures the global slog logger based on the environment
func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // デフォルトはInfo
	}

	// JSONハンドラーを使用
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	logger.Info("ロガーを設定しました", "level", logLevel.String())
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("新しいリクエスト", 
			"method", r.Method, 
			"path", r.URL.Path, 
			"remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}