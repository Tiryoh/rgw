# ADR: dry-runとdescribeコマンドによるエージェント安全機構

- Status: Accepted
- Date: 2026-03-27

## Context

AIエージェントがrgwを操作する場合、以下のリスクがある:

- ハルシネーションしたパラメータでシンボリックリンク操作を実行し、意図しない変更が発生する
- コマンドの引数・フラグ仕様をシステムプロンプトに静的に埋め込むと、トークン浪費かつバージョン不一致が生じる

[Agent-Friendly CLI Design](../references/agent-friendly-cli-design.md) の原則7（dry-run）と原則2（スキーマのランタイムイントロスペクション）に対応が必要だった。

## Decision

グローバル `--dry-run` フラグを追加し変更操作を事前検証可能にする。`rgw describe` コマンドを追加しCobraコマンドツリーからスキーマをJSON出力する。

## Decision Details

### --dry-run

- `root.go` にグローバル `--dry-run` boolフラグを登録
- 対象コマンド: `link set`, `link unset`, `link repair`, `ws add`, `ws use`, `open`
- dry-run時は実行前にメッセージ出力して `return nil`。バリデーション（パス解決、alias算出等）は実行する
- `--output json` との組み合わせで `{"ok":true,"message":"[dry-run] Would link ..."}` を出力
- dry-runガードはコマンドハンドラ内に配置（ライブラリ層には干渉しない）

### describe

- `rgw describe [command...]` で任意のコマンドのメタデータをJSON出力
- Cobraの `cmd.Commands()`, `cmd.LocalFlags()`, `cmd.InheritedFlags()` から動的に構築
- 引数は `Use` 文字列の `<arg>` (必須) および `[arg]` (省略可能) パターンから正規表現で抽出し、括弧の種類で `required` フィールドを決定する
- 静的なスキーマファイルを別途メンテナンスしない

## Alternatives Considered

- **`--dry-run` をライブラリ層（symlink.Set等）に組み込む**: ドメインロジックにUI関心事が混入する。CLIハンドラ層で制御する方が責務が明確
- **`--dry-run` をコマンドローカルフラグにする**: 対象コマンドごとに登録が必要で漏れやすい。グローバルフラグなら一律適用
- **静的JSONスキーマファイルを同梱**: コマンド定義と二重管理になり陳腐化リスクが高い。Cobraから動的生成する方が常に正確
- **OpenAPI/JSON Schema仕様に準拠**: rgwはHTTP APIではないため過剰。独自の軽量スキーマで十分

## Consequences

- エージェントは `--dry-run` で変更操作を事前検証でき、ハルシネーション起因のデータ損失リスクを低減できる
- `describe` によりエージェントはランタイムでCLI仕様を問い合わせ可能になり、システムプロンプトへの静的埋め込みが不要
- `--dry-run` は読み取り専用コマンド（`ws list`, `link status`等）では無意味だが、グローバルフラグのためエラーにはしない（無視される）
- `describe` の引数メタデータ抽出はCobraの `Use` 文字列のヒューリスティックに依存しており、`<arg>` パターンに従わないコマンドでは不正確になる可能性がある

## Verification / Guardrails

- `rgw link set <repo> --dry-run` がシンボリックリンクを作成せずメッセージのみ出力すること
- `rgw describe link set` が引数 `repo`、フラグ `--branch`, `--path`, `--interactive` を含むJSONを出力すること
- `rgw describe` が全サブコマンド一覧を返すこと
- `go build ./...` および `go test ./...` がパスすること
