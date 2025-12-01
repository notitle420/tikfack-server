# TikFack Server Constitution

## Core Principles

### I. Clean Architecture First
業務ロジックは application / domain に閉じ込め、presentation と infrastructure は入出力適応のみに専念する。依存方向は presentation → application → infrastructure（port実装）を厳守し、アダプタ側でドメイン型を直接操作しない。

### II. Transport as Presentation
Connect/gRPC などのハンドラやプレゼンターは presentation 層に配置する。DTO変換・レスポンス組み立てはここで完結させ、infrastructure へは外部サービスや永続化の実装だけを置く。

### III. Testable by Design
ユースケース・アダプタともにモック可能なDIを徹底し、`go test ./...` が常に通る構成を維持する。新規機能はユニットテストまたはモックサーバでの検証を伴い、CI実行を念頭に副作用を排除する。

### IV. Security & Middleware
OIDC/Keycloak 連携や認可はミドルウェア層で統合し、各ハンドラからは一貫した context 情報（user/trace/token）を取得できるようにする。資格情報や環境変数は `.env` 管理とし、コード内にハードコードしない。

### V. Observability & Simplicity
`slog` を用いた構造化ログとトレースID付与を必須とし、外部呼び出しやエラーは全てログ出力する。YAGNI に従い、重複処理はユーティリティや presenter に集約してシンプルさを保つ。

## Engineering Standards
- Go 1.24 toolchain、buf/protoc によるコード生成を必須とする。
- 依存は `go mod tidy` で正規化し、ビルド／テスト前に `scripts/buf-generate.sh` を実行してコード生成する。
- `cmd/app` はコンポジションルートとして依存を束ね、その他の層には new/DI ロジックを置かない。

## Workflow & Quality Gates
- 変更前に計画（Plan tool 等）を共有し、責務境界への影響を明示する。
- PR では lint/format（`gofmt`）済みであること、テスト失敗時は原因をレポートする。
- Connect API 追加時は proto→buf generate→handler実装→テストの順で進める。

## Governance
本憲章は TikFack Server の実装方針を定義し、全PR/レビューはこれへの準拠を確認する。改訂は影響分析と周知を伴うこと。

**Version**: 1.0.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-01
