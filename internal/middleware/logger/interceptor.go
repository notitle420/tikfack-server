package logger

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/bufbuild/connect-go"
	"github.com/tikfack/server/internal/middleware/ctxkeys"
)


func LoggingInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// 1) ctx からユーザーID(sub)を取り出す
			userID, _ := ctx.Value(ctxkeys.SubKey).(string)
			slog.Info("interceptorv2")
			// 2) TraceID を生成し ctx にセット
			traceID := uuid.NewString()
			ctx = context.WithValue(ctx, ctxkeys.TraceIDKey, traceID)

			// 3) ログを出力 (開始)
			slog.Default().With(
				slog.String("trace_id", traceID),
				slog.String("user_id", userID),
			).Info("request started")

			// 4) 次へ
			res, err := next(ctx, req)

			// エラーがあればログを出力
			if err != nil {
				slog.Default().With(
					slog.String("trace_id", traceID),
					slog.String("user_id", userID),
				).Error("request error", "err", err)
				return nil, err
			}

			// 5) ログを出力 (終了)
			slog.Default().With(
				slog.String("trace_id", traceID),
				slog.String("user_id", userID),
			).Info("request completed")

			return res, nil
		}
	})
}
