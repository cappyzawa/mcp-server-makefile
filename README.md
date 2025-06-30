# MCP Server Makefile

Makefile を効率的に探索・解析するための Model Context Protocol (MCP) サーバー

## 概要

MCP Server Makefile は、Claude が Makefile プロジェクトを理解し、ターゲットの依存関係や変数を把握するのを支援する MCP サーバーです。

## 機能

- **ターゲット一覧取得**: Makefile 内のすべてのターゲットを一覧表示
- **ターゲット詳細取得**: 特定ターゲットのコマンドと依存関係を表示
- **依存関係グラフ生成**: ターゲット間の依存関係を可視化
- **変数一覧取得**: Makefile で定義された変数の一覧表示
- **変数展開**: 変数の再帰的展開と解決
- **Makefile 検索**: プロジェクト内のすべての Makefile を検索

## インストール

### 方法1: go install を使用（推奨）

```bash
go install github.com/cappyzawa/mcp-server-makefile@latest
```

これにより、バイナリが `$GOPATH/bin` にインストールされます。

### 方法2: ソースからビルド

```bash
git clone https://github.com/cappyzawa/mcp-server-makefile.git
cd mcp-server-makefile
go build -o mcp-server-makefile
```

### Claude MCP での登録

`claude_desktop_config.json` に以下を追加：

```json
{
  "mcpServers": {
    "makefile": {
      "command": "mcp-server-makefile"
    }
  }
}
```

注: `go install` でインストールした場合、`$GOPATH/bin` が PATH に含まれていれば、フルパスを指定する必要はありません。

または `claude mcp` コマンドを使用：

```bash
# go install でインストールした場合
claude mcp add makefile mcp-server-makefile

# フルパスを指定する場合
claude mcp add makefile $GOPATH/bin/mcp-server-makefile
```

## 使用方法

Claude で以下のようなコマンドを実行できます：

### ターゲット一覧の確認
```
list_targets で Makefile のターゲットを確認します
```

### 特定ターゲットの詳細
```
get_target で build ターゲットの詳細を確認します
```

### 依存関係の確認
```
get_dependencies で test ターゲットの依存関係を確認します
```

### 変数の一覧
```
list_variables で定義されている変数を確認します
```

### 変数の展開
```
expand_variable で CFLAGS 変数を展開します
```

### Makefile の検索
```
find_makefiles でプロジェクト内のすべての Makefile を検索します
```

## 開発

### 必要な環境

- Go 1.21 以上

### テストの実行

```bash
go test ./...
```

### コードの品質チェック

```bash
go vet ./...
golangci-lint run
```

### リリース

新しいバージョンをリリースする際は、Git タグを作成してプッシュしてください：

```bash
git tag v1.0.0
git push origin v1.0.0
```

これにより、`go install github.com/cappyzawa/mcp-server-makefile@v1.0.0` のように特定のバージョンをインストールできるようになります。

## ライセンス

MIT License

## 貢献

Issue や Pull Request を歓迎します。

1. Fork してください
2. Feature ブランチを作成してください (`git checkout -b feature/amazing-feature`)
3. 変更をコミットしてください (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュしてください (`git push origin feature/amazing-feature`)
5. Pull Request を開いてください