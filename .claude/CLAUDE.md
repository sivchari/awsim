# awsim 開発ガイドライン

## 重要ルール（必ず守ること）

### PR 分割ルール

**1 Issue = 1 PR を厳守する.**

- 複数の Issue を 1 つの PR にまとめない
- 複数の機能を 1 つの PR に含めない
- 基盤変更（interface, storage, server）と機能追加（S3, SQS 等）は別 PR
- ドキュメント変更と実装変更は別 PR

### PR 作成前チェック

PR を作成する前に以下を確認:

1. **単一責任**: この PR は 1 つの Issue のみを解決しているか？
2. **スコープ**: 関係ない変更が混入していないか？
3. **テスト**: 対象機能のテストのみ追加されているか？

### 禁止事項

- 「ついでに」別の機能を追加する
- Issue なしで機能を実装する
- 複数サービスの実装を 1 PR にまとめる

### 例外

以下の場合のみ複数ファイルの変更を 1 PR に含めることができる:

- 1 つの Issue を解決するために必要な関連変更
- リファクタリングで複数ファイルに影響する場合（専用 Issue を作成）

## プロジェクト概要

awsim は CI/CD 環境向けの軽量 AWS サービスエミュレータです.

### 特徴

- **認証不要**: AWS 認証情報なしで動作
- **高速起動**: コンテナ起動から数秒で利用可能
- **AWS SDK v2 互換**: Go の AWS SDK v2 と完全互換
- **軽量**: 単一バイナリ、最小限のリソース消費

### 対象ユーザー

- CI/CD パイプラインで AWS サービスをテストする開発者
- ローカル開発環境で AWS サービスをモックしたい開発者

## アーキテクチャ

### ディレクトリ構造

```
awsim/
├── cmd/awsim/
│   └── main.go                    # エントリーポイント
├── internal/
│   ├── server/                    # HTTP サーバー & ルーティング
│   │   ├── server.go              # サーバー設定 & 起動
│   │   └── router.go              # サービスルーティング
│   ├── service/                   # サービス実装
│   │   ├── interface.go           # Service interface 定義
│   │   ├── registry.go            # Service registry
│   │   ├── s3/                    # S3 サービス
│   │   │   ├── service.go         # サービス登録
│   │   │   ├── handlers.go        # オペレーションハンドラ
│   │   │   ├── types.go           # Request/Response 型
│   │   │   └── storage.go         # S3 固有ストレージロジック
│   │   ├── sqs/                   # SQS サービス
│   │   ├── dynamodb/              # DynamoDB サービス
│   │   └── ...                    # 他サービス (1 サービス = 1 ディレクトリ)
│   ├── storage/                   # ストレージバックエンド
│   │   ├── interface.go           # Storage interface
│   │   └── memory.go              # インメモリ実装
│   ├── protocol/                  # プロトコルハンドラ
│   │   ├── restjson.go            # REST JSON (S3, Secrets Manager)
│   │   ├── query.go               # Query protocol (SQS)
│   │   └── awsjson.go             # AWS JSON (DynamoDB)
│   └── errors/                    # エラー定義
│       └── errors.go              # AWS 互換エラー
├── test/
│   ├── integration/               # Integration tests (AWS SDK v2)
│   └── benchmark/                 # パフォーマンスベンチマーク
└── tools/
    └── go.mod                     # ツール依存関係
```

### 設計原則

1. **レジストリベースのサービス発見**: 各サービスは起動時に自動登録
2. **インターフェースベースの設計**: 共通インターフェースで一貫性を保証
3. **プロトコル抽象化**: REST JSON, Query, AWS JSON を統一的に処理
4. **メモリファーストストレージ**: CI/CD に最適化、永続化不要

## サービス実装パターン

### Service Interface

```go
// Service は AWS サービスの共通インターフェース
type Service interface {
    // Name はサービス名を返す (e.g., "s3", "sqs", "dynamodb")
    Name() string

    // Prefix は URL プレフィックスを返す (e.g., "/s3")
    // Host ベースのルーティングの場合は空文字列
    Prefix() string

    // RegisterRoutes はルートを登録する
    RegisterRoutes(r Router)
}
```

### Handler Pattern

```go
// ハンドラの基本構造
func (s *S3Service) PutObject(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. リクエストパース
    bucket, key := parseBucketKey(r.URL.Path)
    body, err := io.ReadAll(r.Body)
    if err != nil {
        writeError(w, errors.NewInternalError(err))
        return
    }

    // 2. バリデーション
    if bucket == "" {
        writeError(w, errors.NewInvalidBucketName())
        return
    }

    // 3. オペレーション実行
    if err := s.storage.PutObject(ctx, bucket, key, body); err != nil {
        writeError(w, err)
        return
    }

    // 4. AWS 互換レスポンス返却
    w.Header().Set("ETag", calculateETag(body))
    w.WriteHeader(http.StatusOK)
}
```

### エラーハンドリング

```go
// AWS 互換エラーレスポンス
type AWSError struct {
    Code       string `xml:"Code"`
    Message    string `xml:"Message"`
    Resource   string `xml:"Resource,omitempty"`
    RequestID  string `xml:"RequestId"`
}

// エラー返却
func writeError(w http.ResponseWriter, err *AWSError) {
    w.Header().Set("Content-Type", "application/xml")
    w.WriteHeader(err.HTTPStatus())
    xml.NewEncoder(w).Encode(err)
}
```

## 開発ワークフロー

### Issue 駆動開発

**全ての実装は Issue から始まる.**

1. 実装前に必ず対応する Issue を確認
2. Issue がなければ先に Issue を作成
3. 1 つの Issue に対して 1 つの PR を作成

### 新機能実装の流れ

1. **Issue 確認**: GitHub Issues から実装対象を **1 つだけ** 選択
2. **Feature branch 作成**:
   ```bash
   git checkout -b feat/{service}-{operation}
   # 例: git checkout -b feat/s3-put-object
   ```
3. **その Issue のみ実装**: スコープ外の変更は別 PR
4. **Integration test 追加**: AWS SDK v2 を使用したテスト
5. **Lint 実行**:
   ```bash
   make lint
   ```
6. **Test 実行**:
   ```bash
   make test
   ```
7. **PR 作成**: **1 Issue = 1 PR** で Issue を参照

### PR 作成時の確認事項

```bash
# 変更ファイルを確認
git diff --name-only main

# 変更が対象 Issue のスコープ内か確認
# スコープ外の変更があれば別 PR に分ける
```

### ブランチ命名規則

| プレフィックス | 用途 | 例 |
|---------------|------|-----|
| `feat/` | 新機能 | `feat/s3-put-object` |
| `fix/` | バグ修正 | `fix/s3-etag-calculation` |
| `refactor/` | リファクタリング | `refactor/storage-interface` |
| `test/` | テスト追加 | `test/sqs-integration` |
| `docs/` | ドキュメント | `docs/readme-update` |

## テスト戦略

### テストの種類

1. **Unit tests**: 内部ロジックのテスト
   - ストレージ操作
   - プロトコルパース
   - エラー生成

2. **Integration tests**: AWS SDK v2 との互換性テスト
   - 実際の SDK クライアントを使用
   - エンドツーエンドの動作確認

3. **Benchmark tests**: パフォーマンス測定
   - レイテンシ
   - スループット
   - メモリ使用量

### Integration Test の書き方

```go
func TestS3_PutObject(t *testing.T) {
    // 1. awsim サーバー起動
    srv := startTestServer(t)
    defer srv.Close()

    // 2. AWS SDK v2 クライアント作成
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            "test", "test", "",
        )),
    )
    require.NoError(t, err)

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.BaseEndpoint = aws.String(srv.URL)
        o.UsePathStyle = true
    })

    // 3. オペレーション実行
    _, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
        Bucket: aws.String("test-bucket"),
    })
    require.NoError(t, err)

    _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String("test-bucket"),
        Key:    aws.String("test-key"),
        Body:   strings.NewReader("test content"),
    })
    require.NoError(t, err)

    // 4. 検証
    result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
        Bucket: aws.String("test-bucket"),
        Key:    aws.String("test-key"),
    })
    require.NoError(t, err)

    body, _ := io.ReadAll(result.Body)
    assert.Equal(t, "test content", string(body))
}
```

### テスト実行コマンド

```bash
# 全テスト実行
make test

# Integration test のみ
make test-integration

# Benchmark
make bench
```

## コード品質

### Linter ルール

- `golangci-lint` の全ルールをパス
- `.golangci.yaml` で設定済み

### コーディング規約

1. **エクスポート関数**: ドキュメントコメント必須
   ```go
   // PutObject stores an object in the specified bucket.
   func (s *S3Service) PutObject(...) { ... }
   ```

2. **エラーハンドリング**: wrapped errors を使用
   ```go
   if err != nil {
       return fmt.Errorf("failed to put object: %w", err)
   }
   ```

3. **Context 伝播**: 全ての関数で context を受け取る
   ```go
   func (s *Storage) PutObject(ctx context.Context, ...) error
   ```

4. **命名規則**:
   - インターフェース: 動詞 + er (e.g., `Reader`, `Writer`)
   - 構造体: 名詞 (e.g., `S3Service`, `MemoryStorage`)
   - 関数: 動詞で始める (e.g., `CreateBucket`, `PutObject`)

## コミット規約

### コミットメッセージ形式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

| Type | 説明 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `refactor` | リファクタリング |
| `test` | テスト追加/修正 |
| `docs` | ドキュメント |
| `chore` | その他 (CI, ビルド等) |

### Scope

サービス名または機能名: `s3`, `sqs`, `dynamodb`, `server`, `storage`

### 例

```
feat(s3): add PutObject operation

- Implement PutObject handler with multipart support
- Add ETag calculation
- Add integration tests with AWS SDK v2

Closes #1
```

## サービス実装チェックリスト

新しいサービスを実装する際のチェックリスト:

- [ ] `internal/service/{service}/` ディレクトリ作成
- [ ] `service.go`: Service interface 実装
- [ ] `handlers.go`: オペレーションハンドラ実装
- [ ] `types.go`: Request/Response 型定義
- [ ] `storage.go`: サービス固有のストレージロジック (必要な場合)
- [ ] `internal/service/registry.go` に登録
- [ ] `test/integration/{service}_test.go`: Integration tests
- [ ] README.md 更新 (サポートサービス一覧)

## AWS 互換性ルール

### レスポンス形式

- XML レスポンス: S3, SQS (Query protocol)
- JSON レスポンス: DynamoDB, Secrets Manager

### エラーコード

AWS 公式ドキュメントに準拠:
- S3: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
- SQS: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/CommonErrors.html
- DynamoDB: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.Errors.html

### 必須ヘッダー

```
x-amz-request-id: {unique-id}
x-amz-id-2: {secondary-id}  # S3 only
Content-Type: application/xml or application/json
```

## 環境変数

| 変数名 | デフォルト | 説明 |
|--------|-----------|------|
| `AWSIM_HOST` | `0.0.0.0` | バインドホスト |
| `AWSIM_PORT` | `4566` | ポート番号 |
| `AWSIM_LOG_LEVEL` | `info` | ログレベル (debug, info, warn, error) |
| `AWSIM_SERVICES` | (all) | 有効化するサービス (カンマ区切り) |

## Makefile ターゲット

```bash
make build          # バイナリビルド
make test           # 全テスト実行
make test-unit      # Unit test のみ
make test-integration # Integration test のみ
make bench          # ベンチマーク
make lint           # Linter 実行
make lint-fix       # Linter 自動修正
make docker-build   # Docker イメージビルド
make docker-run     # Docker コンテナ起動
```

## FAQ

### Q: 新しいオペレーションを追加するには？

1. 対応する Issue を確認/作成
2. `internal/service/{service}/handlers.go` にハンドラ追加
3. `service.go` の `RegisterRoutes` にルート追加
4. Integration test 追加
5. PR 作成

### Q: 新しいサービスを追加するには？

1. 対応する Issue を確認/作成
2. `internal/service/{service}/` ディレクトリ作成
3. チェックリストに従って実装
4. `internal/service/registry.go` に登録
5. Integration test 追加
6. README.md 更新
7. PR 作成

### Q: エラーレスポンスの形式は？

AWS 公式と同じ形式を使用:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Error>
    <Code>NoSuchBucket</Code>
    <Message>The specified bucket does not exist</Message>
    <Resource>/bucket-name</Resource>
    <RequestId>xxx</RequestId>
</Error>
```
