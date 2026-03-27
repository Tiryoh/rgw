# ADR: 境界での入力バリデーションによるハルシネーション対策

- Status: Accepted
- Date: 2026-03-27

## Context

rgw はファイルシステム操作（シンボリックリンク作成・削除）を行うCLIであり、AIエージェントがハルシネーションした入力を渡した場合にセキュリティリスクが生じる。具体的な脅威:

- alias名に `../../.ssh` を含めてワークスペース外にシンボリックリンクを作成（パストラバーサル）
- `\x00` や `\n` を含むリポジトリ名でファイルシステムを汚染（制御文字）
- `src_subdir` に `../` を設定してワークスペース外を操作（設定ファイル経由の攻撃）

[Agent-Friendly CLI Design](../references/agent-friendly-cli-design.md) の原則4（ハルシネーション対策の入力バリデーション）に従い、CLIが最終防御線として機能する必要がある。

## Decision

`internal/validate` パッケージを新設し、外部入力が内部データに変換される境界でバリデーションを行う。

## Decision Details

4つの純粋関数を提供する:

| 関数 | 用途 |
|------|------|
| `SafePath(base, child)` | `filepath.Clean` 後に child が base 内に収まるか検証 |
| `NoControlChars(s)` | ASCII 0x20 未満のバイトを拒否 |
| `WorkspaceName(name)` | `[a-zA-Z0-9_-]{1,128}` のホワイトリスト |
| `RepoSegment(s)` | `..`, `?`, `#`, `%`, 制御文字を拒否 |

適用箇所:

| 境界 | バリデーション |
|------|----------------|
| `symlink.Set/Unset` | `SafePath` でalias検証 |
| `ghq.ParseRepoArg` | `RepoSegment` で各セグメント検証 |
| `workspace.Add` | `WorkspaceName` + `NoControlChars` |
| `config.Load` | `SafePath` でSrcSubdir検証 |
| `alias.Resolve` | `NoControlChars` で結合結果検証 |

`alias.Resolve` はシグネチャを `string` → `(string, error)` に変更し、バリデーション失敗を呼び出し元に伝播する。

内部データ（git worktree出力等）は信頼し、バリデーション対象外とする。

## Alternatives Considered

- **CLI層（コマンドハンドラ）でバリデーション**: コマンドごとに重複するバリデーションコードが発生する
- **ドメイン層の各関数内でバリデーション**: バリデーションロジックが各パッケージに分散し、一貫性を保ちにくい
- **`alias.Resolve` のシグネチャを維持し、入力側のみで検証**: 結合結果の安全性を保証できない

## Consequences

- バリデーションが純粋関数として独立しており、テストが容易
- `SafePath` は `filepath.Abs` + prefix チェックにより、`filepath.Join` の正規化だけでは防げない攻撃を防御
- `alias.Resolve` のシグネチャ変更は破壊的変更であり、全呼び出し元の修正が必要だった
- `WorkspaceName` のホワイトリストは厳格（ドットや日本語名は不可）。ROS ワークスペース名の慣習に沿った制約だが、将来緩和が必要になる可能性がある

## Verification / Guardrails

- `internal/validate/validate_test.go` でパストラバーサル・制御文字・不正名の拒否をテーブル駆動テストで検証
- `go test ./...` が全パスすること
- `../` を含むリポジトリ引数・alias名がエラーで拒否されること
- 既存の正当な入力（`Tiryoh__my_pkg` 等）が引き続き受け入れられること
