# MCP Server Makefile 仕様書

## 概要

MCP Server Makefile は、Makefile の効率的な探索と解析を可能にする Model Context Protocol (MCP) サーバーです。
Claude が Makefile プロジェクトを理解し、ターゲットの依存関係や変数を把握するのを支援します。

## 主要機能

### 1. ターゲット一覧取得 (list_targets)

- Makefile 内のすべてのターゲットを一覧表示
- ターゲットの説明（コメント）も含む
- PHONY ターゲットの識別

### 2. ターゲット詳細取得 (get_target)

- 特定ターゲットのコマンド一覧
- 依存関係の展開
- 関連する変数の表示

### 3. 依存関係グラフ生成 (get_dependencies)

- ターゲット間の依存関係を可視化
- 循環依存の検出
- 依存関係の深さ制限オプション

### 4. 変数一覧取得 (list_variables)

- Makefile で定義された変数の一覧
- 変数の値と定義位置
- 環境変数との区別

### 5. 変数展開 (expand_variable)

- 変数の再帰的展開
- 条件付き代入の解決
- シェル変数との統合

### 6. Makefile 検索 (find_makefiles)

- プロジェクト内のすべての Makefile を検索
- include されるファイルの追跡
- カスタムパターンのサポート

## 実装の詳細

### パーサー設計

- GNU Make の構文をサポート
- 継続行（バックスラッシュ）の処理
- 条件文（ifeq, ifdef など）の解析
- include ディレクティブの処理

### エラーハンドリング

- 構文エラーの詳細な報告
- 循環依存の検出と報告
- 未定義変数の警告

### パフォーマンス最適化

- Makefile のキャッシュ機構
- 大規模プロジェクトでの効率的な探索
- 並列処理による高速化

## API 仕様

### Tools

#### list_targets

```json
{
  "name": "list_targets",
  "description": "List all targets in the Makefile",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Path to the Makefile (optional, defaults to ./Makefile)"
      }
    }
  }
}
```

#### get_target

```json
{
  "name": "get_target",
  "description": "Get detailed information about a specific target",
  "inputSchema": {
    "type": "object",
    "properties": {
      "target": {
        "type": "string",
        "description": "Target name"
      },
      "path": {
        "type": "string",
        "description": "Path to the Makefile (optional)"
      }
    },
    "required": ["target"]
  }
}
```

#### get_dependencies

```json
{
  "name": "get_dependencies",
  "description": "Get dependency graph for a target",
  "inputSchema": {
    "type": "object",
    "properties": {
      "target": {
        "type": "string",
        "description": "Target name"
      },
      "path": {
        "type": "string",
        "description": "Path to the Makefile (optional)"
      },
      "max_depth": {
        "type": "integer",
        "description": "Maximum dependency depth (optional)"
      }
    },
    "required": ["target"]
  }
}
```

#### list_variables

```json
{
  "name": "list_variables",
  "description": "List all variables defined in the Makefile",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Path to the Makefile (optional)"
      },
      "include_env": {
        "type": "boolean",
        "description": "Include environment variables (default: false)"
      }
    }
  }
}
```

#### expand_variable

```json
{
  "name": "expand_variable",
  "description": "Expand a variable to its full value",
  "inputSchema": {
    "type": "object",
    "properties": {
      "variable": {
        "type": "string",
        "description": "Variable name"
      },
      "path": {
        "type": "string",
        "description": "Path to the Makefile (optional)"
      }
    },
    "required": ["variable"]
  }
}
```

#### find_makefiles

```json
{
  "name": "find_makefiles",
  "description": "Find all Makefiles in the project",
  "inputSchema": {
    "type": "object",
    "properties": {
      "root": {
        "type": "string",
        "description": "Root directory to search (optional, defaults to current directory)"
      },
      "pattern": {
        "type": "string",
        "description": "File pattern to match (optional, defaults to common Makefile names)"
      }
    }
  }
}
```

## 使用例

### ターゲット一覧の取得

```
Claude: list_targets で Makefile のターゲットを確認します
結果: build, test, clean, install, docker-build
```

### 依存関係の確認

```
Claude: get_dependencies で build ターゲットの依存関係を確認します
結果: build -> compile -> [generate, vendor]
```

### 変数の展開

```
Claude: expand_variable で CFLAGS 変数を展開します
結果: "-O2 -Wall -Wextra -std=c11 -I./include"
```

