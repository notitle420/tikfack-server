# TikFack Server

TikFack サーバーは、動画配信のバックエンドを担当する Go で書かれたアプリケーションです。このサーバーは Connect プロトコルを使用して API を提供し、DMM API を利用して動画データを取得します。

クリーンアーキテクチャの原則に従って設計され、高い保守性とテスタビリティを実現しています。

## 主な機能

### コア機能
- **DMM API 統合**: DMM の動画データベースからリアルタイムで動画情報を取得
- **Connect プロトコル対応**: gRPC と REST の両方をサポートする統一API
- **動画検索**: 日付、キーワード、IDによる柔軟な動画検索機能
- **認証・認可**: OIDC と Keycloak を使用したセキュアな認証システム

### アーキテクチャ特徴
- **クリーンアーキテクチャ**: ドメイン駆動設計による明確な責務分離
- **依存性注入**: 高いテスタビリティと拡張性を実現
- **包括的テスト**: ユニットテストとモックを活用した高品質コード
- **ミドルウェア対応**: ログ記録、認証、CORS などの横断的関心事を効率的に処理

## 必要条件

### システム要件
- **Go**: 1.23.0 以上（toolchain: go1.24.3）
- **Protocol Buffers**: protoc コンパイラ
- **buf**: Protocol Buffers 生成ツール
- **Docker**: コンテナ実行用（オプション）

### 外部サービス
- **DMM API アカウント**: API ID とアフィリエイト ID が必要
- **認証プロバイダー**: Keycloak

## 環境変数

プロジェクトのルートディレクトリに `.env` ファイルを作成し、以下の環境変数を設定してください：

### 必須設定
```bash

DMM_API_ID= #DMMのAPIID
DMM_API_AFFILIATE_ID= #DMMのアフィリエイトID
PORT=50051 #50051でOK
BASE_URL=https://api.dmm.com/affiliate/ #DMMのベースURL
HITS=10 # オプション: DMM API から取得する動画の数
LOG_LEVEL=debug #ログレベル
ISSUER_URL= #keycloakのベースURL http://localhost:18080/realms/myrealm
CLIENT_ID= #keycloackのバックエンド用のクライアントID
KEYCLOAK_REALM= #keycloackのREALM 
KEYCLOAK_BACKEND_CLIENT_SECRET= #keycloackのバックエンド用のクライアントID keycloak → ckuebts → backuendclients → credentials
KEYCLOAK_BASE_URL= #keycloakのベースURL：例 http://localhost:18080
KAFKA_BROKER_ADDRESSES=localhost:9094 # カンマ区切りでKafkaブローカーのホスト:ポートを設定（未設定時はlocalhost:9094）
```

## Protocol Buffers のセットアップ

サーバーを実行する前に、Protocol Buffers からコードを生成する必要があります：

```bash
# buf ツールのインストール（必要な場合）
go install github.com/bufbuild/buf/cmd/buf@latest

# Protocol Buffers コードの生成
buf generate
```

これにより、`gen` ディレクトリに必要なコードが生成されます。

## インストールと実行

### ローカル開発環境

```bash
# 1. リポジトリのクローン
git clone https://github.com/tikfack/server.git
cd tikfack-server

# 2. 依存関係のインストール
go mod download

# 3. 環境変数の設定
cp .env.example .env  # .env ファイルを作成し、必要な値を設定

# 4. Protocol Buffers コードの生成
buf generate

# 5. サーバーの起動
go run cmd/app/main.go
```

gRPCは `localhost:50051` でアクセス可能です。

### Docker を使った実行

```bash
# Docker イメージのビルド
docker build -t tikfack-server .

# コンテナの実行
docker run -p 50051:50051 --env-file .env tikfack-server
```

## トラブルシューティング

### パッケージのインポートエラー

以下のようなエラーが表示される場合：

```
package server/generated is not in std
package server/generated/protoconnect is not in std
```

Protocol Buffers のコードが正しく生成されていない可能性があります。以下の手順を試してください：

1. `buf generate` コマンドを実行して Protocol Buffers コードを生成する
2. `go mod tidy` を実行して依存関係を更新する
3. もし問題が解決しない場合は、`go.mod` ファイルの `module` 行が正しいパスになっているか確認する

## API エンドポイント

サーバーは以下の API エンドポイントを提供します：

### VideoService API

#### GetVideos
動画のリストを取得します。
- **URL**: `/proto.video.v1.VideoService/GetVideos`
- **メソッド**: POST（Connect プロトコル）
- **機能**: 指定した条件に基づいて動画一覧を取得

#### GetVideoById
指定した ID の動画情報を取得します。
- **URL**: `/proto.video.v1.VideoService/GetVideoById`
- **メソッド**: POST（Connect プロトコル）
- **パラメーター**: 動画ID
- **レスポンス**: 動画の詳細情報

#### SearchVideos
キーワードによる動画検索を行います。
- **URL**: `/proto.video.v1.VideoService/SearchVideos`
- **メソッド**: POST（Connect プロトコル）
- **機能**: タイトル、出演者、ジャンル等での横断検索

#### GetVideosByDate
日付範囲による動画検索を行います。
- **URL**: `/proto.video.v1.VideoService/GetVideosByDate`
- **メソッド**: POST（Connect プロトコル）
- **機能**: 指定期間内にリリースされた動画を取得

#### GetVideosByKeyword
キーワード検索による動画取得を行います。
- **URL**: `/proto.video.v1.VideoService/GetVideosByKeyword`
- **メソッド**: POST（Connect プロトコル）
- **機能**: 特定キーワードに一致する動画を検索

## プロジェクト構造

プロジェクトはクリーンアーキテクチャに従って構成されています：

```
tikfack-server/
├── cmd/
│   └── app/              # アプリケーションエントリポイント
│       └── main.go       # メインプログラム
├── internal/             # 内部パッケージ
│   ├── application/      # アプリケーション層
│   │   └── usecase/      # ビジネスロジック
│   │       └── video/    # 動画関連ユースケース
│   ├── domain/           # ドメイン層
│   │   ├── entity/       # ドメインエンティティ
│   │   └── repository/   # リポジトリインターフェース
│   ├── infrastructure/   # インフラストラクチャ層
│   │   ├── connect/      # Connect API ハンドラー
│   │   ├── dmmapi/       # DMM API クライアント
│   │   ├── repository/   # リポジトリ実装
│   │   └── util/         # ユーティリティ
│   └── middleware/       # ミドルウェア
│       ├── auth/         # 認証機能
│       ├── logger/       # ロギング
│       └── ctxkeys/      # コンテキストキー
├── proto/                # Protocol Buffers 定義
├── gen/                  # 生成されたコード (Protocol Buffers)
├── test/                 # テスト関連ファイル
├── docs/                 # ドキュメント
├── .github/              # GitHub Actions設定
├── Dockerfile            # Docker ビルド設定
├── buf.gen.yaml          # buf 生成設定
├── go.mod               # Go モジュール定義
├── go.sum               # 依存関係のハッシュ
└── README.md            # このファイル
```

### アーキテクチャレイヤー

- **Application Layer**: ビジネスロジックとユースケース
- **Domain Layer**: ドメインエンティティとビジネスルール
- **Infrastructure Layer**: 外部サービスとの統合、データアクセス
- **Middleware**: 横断的関心事（認証、ログ、CORS等）

## 開発とテスト

### テストの実行

```bash
# 全テストの実行
go test ./...

# テストカバレッジの確認
go test -cover ./...

# 特定パッケージのテスト
go test ./internal/application/usecase/video/...
```

### コード品質

プロジェクトには包括的なテストスイートが含まれています：
- **ユニットテスト**: 各レイヤーの個別機能をテスト
- **モック**: `go.uber.org/mock` を使用した依存関係のモック化
- **テーブル駆動テスト**: 複数のテストケースを効率的に実行

### デバッグとログ

- 構造化ログ出力による詳細なトレーシング
- リクエスト/レスポンスの自動ログ記録
- エラー詳細の包括的な記録
