# AGENTS.md

## Project Overview
- Go 製の動画配信バックエンド。DMM/FANZA API から動画カタログを取得し、Connect 経由で gRPC / gRPC-Web / REST を同時に公開する。
- バウンデッドコンテキスト: `VideoCatalog` (DMM 連携・検索), `EventLog` (再生・行動ログを Kafka へ送信), `Auth` (Keycloak OIDC による認証・権限チェック)。
- 主要アグリゲート: `User` (Keycloak ID を主体としたユーザー), `EventLog` (動画視聴イベント)。Video は外部カタログからの DTO として扱う。

## Architecture & Folder Structure
- クリーンアーキテクチャ準拠の4層。
  - `internal/domain`: エンティティ (`entity/User`, `entity/EventLog`) とリポジトリIF (`repository/*`)。外部依存を持たない。
  - `internal/application`: ユースケース (`usecase/video`, `usecase/event_log`)、ポート (`port/VideoCatalog`)、DTO (`model/*`)。ドメインを操作し、インフラにはポート越しに依存。
  - `internal/infrastructure`: ポート実装 (DMM API クライアント `dmmapi`, Kafka リポジトリ `repository/event_log`)、ユーティリティ。外部サービスとの通信のみを担当。
  - `internal/presentation/connect`: Connect ハンドラとプレゼンター。proto/gen との変換とバリデーション、ミドルウェア適用。
- DI: `internal/di` で Google Wire を利用。`InitializeVideoHandler` は `NewVideoRepository -> NewVideoUsecase -> NewVideoServiceHandler` を組み立て。`InitializeEventLogHandler` は `provideKafkaWriter -> NewKafkaEventLogRepository -> NewEventLogService -> NewEventLogServiceHandler` を組み立て。ハンドラ生成時に Connect のインターセプタ配列を受け取り、`cmd/app/main.go` から注入する。
- 横断関心: `internal/middleware` (auth・logger・ctxkeys)、`internal/di` (Wire による依存性注入)。
- 境界ルール: ドメインはインフラを知らない。ユースケースはポート経由で依存注入する。プレゼンテーションはユースケースの戻り値のみを扱い、インフラ型を直接参照しない。

## Coding Style & Domain Modeling Guidelines
- 命名: エンティティ/値は PascalCase (`User`, `EventLog`)、リポジトリは `XRepository`、ユースケースは `XUsecase`。ポートは外部境界を表す名前 (`VideoCatalog`)。エラーは `ErrXxx`。
- ドメインロジック配置: エンティティの整合性はエンティティメソッドで担保 (`User.Validate`, `AddFolderID`, `UpdateEmail` など)。ユースケースでは入出力の正規化 (hits/offset の clamp、日付 parse) とポート呼び出しを行う。
- 不変条件/集約ルール: `User` のアカウント名は3–32文字、メールは簡易形式チェック、KeycloakID は必須。フォルダIDは重複不可で更新時に `UpdatedAt` を更新。`EventLog` は UUID/ID と event_time の組で Kafka キー化される前提。
- 実装指針: コンテキストから `ctxkeys` でトレース/ユーザー情報を取り出し slog に付与。外部 API 呼び出しは `application/port` を経由し、ハンドラから直接インフラ実装を参照しない。

## Build / Test / Run Commands
- Go: `go 1.23.0` (`toolchain go1.24.3`)。
- 起動: `go run cmd/app/main.go`。Docker: `docker compose up --build` (env は `docker.env` / `.env`)。
- ビルド: `go build ./...`。
- テスト: `go test ./...`。カバレッジ: `go test -coverprofile=coverage.out ./...` → `go tool cover -html=coverage.out -o coverage.html`。
- Proto 生成: `./scripts/buf-generate.sh [--path proto/event_log]` (Docker 必須)。
- モック/DI 生成: `go generate ./...` (mockgen, wire)。Wire のみ再生成する場合は `cd internal/di && go generate ./...`。

## CI / Release / Deployment
- CI 定義はリポジトリ内に未配置 (2025-XX 時点)。ローカルで lint/test を実行してからデプロイする前提。
- ブランチ/PR 戦略は明示なし。PR テンプレート `.github/PULL_REQUEST_TEMPLATE.md` に準拠。
- 環境変数 (主要): `DMM_API_ID`, `DMM_API_AFFILIATE_ID`, `ISSUER_URL`, `CLIENT_ID`, `KEYCLOAK_REALM`, `KEYCLOAK_BACKEND_CLIENT_SECRET`, `KEYCLOAK_BASE_URL`, `KAFKA_BROKER_ADDRESSES`, `PORT`, `LOG_LEVEL`。`godotenv` が `.env` を自動読込、Docker は `docker.env`。
- デプロイ/運用: 外部ネットワーク `backend` で Keycloak/Kafka へ到達できることを確認。Kafka の topic は `event-logs` (WriterConfig) を使用。Connect/gRPC/REST は同一ポート `:50051`。

## Glossary / Domain Terms
- DMM API / FANZA: 動画カタログ提供元。パラメータ例 `site=FANZA`, `service=digital`, `floor=videoa`。
- Connect: buf の Connect プロトコル。`videoconnect.NewVideoServiceHandler` / `event_logconnect.NewEventLogServiceHandler` を使用。
- VideoService: 動画検索 API 群 (`GetVideosByDate/Id`, `SearchVideos`, `GetVideosByID`, `GetVideosByKeyword`)。
- EventLogService: 視聴イベントを受け取り Kafka に送る (`Record`, `RecordBatch`)。
- Keycloak / OIDC: Token 検証 (`IntrospectionInterceptor`)、権限確認 (`PermissionInterceptor`, RPT) を行う IdP。
- Naming: Go は PascalCase/export、小文字で非公開。proto フィールドは snake_case。環境変数は大文字スネークケース。

## Restrictions / What Not to Do
- ドメイン層にインフラ依存 (Kafka クライアント、HTTP クライアント等) を持ち込まない。インフラはポート実装に限定。
- ハンドラから直接 DMM/Kafka へアクセスしない。必ずユースケース/ポート経由で呼び出す。
- 認証・権限制御をバイパスしない。Connect ハンドラ追加時は `IntrospectionInterceptor` / `PermissionInterceptor` / `LoggingInterceptor` を組み込む。
- コンテキストを捨てない (log/外部呼び出しは ctx 付きで)。トークン/ユーザーID をログに丸出ししない (token は logger 側で短縮される前提)。
- エンティティの不変条件をスキップして値を直接改変しない (例: `User` 更新時は `Validate` を通す)。
