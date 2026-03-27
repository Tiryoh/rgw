# ADR: JSON出力とTTY自動検出によるデフォルト切替

- Status: Accepted
- Date: 2026-03-27

## Context

rgw はテキスト（tabwriter）出力のみだったため、AIエージェントが出力をパースするにはスクレイピングが必要だった。[Agent-Friendly CLI Design](../references/agent-friendly-cli-design.md) の原則1（生JSONペイロード）に対応するため、構造化出力が必要になった。

デフォルト値の決め方にも複数の選択肢があり、人間とエージェントの両方が自然に使えるデフォルトを選ぶ必要があった。

## Decision

`--output` (`-o`) グローバルフラグを `text` / `json` の2値で導入し、明示指定がない場合はstdoutのTTY判定で自動切替する（TTY → text、非TTY → json）。

## Decision Details

- 実装は `internal/cli/output.go` に集約。`isJSON()` / `printJSON()` / `printAction()` ヘルパーで各コマンドのJSON分岐を最小コードで実現する
- TTY検出には `mattn/go-isatty`（既にindirect依存だったものをdirectに昇格）を使用
- データ出力コマンド（`ws list`, `wt list`, `link status`, `doctor`）は対応する構造体を直接JSONシリアライズする
- アクションコマンド（`link set`, `ws add` 等）は `actionResult{OK bool, Message string}` で統一的にJSON出力する
- 空リストはテキストモードではフレンドリーメッセージ、JSONモードでは `[]` を返す
- 構造体には `json:"..."` タグを付与。`doctor.Severity` は `MarshalText()` を実装し文字列としてシリアライズする

## Alternatives Considered

- **`--json` boolフラグ**: シンプルだが将来 yaml/csv 等への拡張性がない
- **環境変数 `OUTPUT_FORMAT` のみ**: フラグとの併用は可能だが、単体では発見しにくい
- **常にテキストデフォルト（明示 `--output json` 必須）**: エージェントがパイプライン利用するたびにフラグ指定が必要で冗長

## Consequences

- パイプライン（`rgw ws list | jq ...`）やエージェント利用時に自動的にJSONが出力される
- `--output` は将来 `yaml` 等に拡張可能
- TTY検出は一部環境（CI内のリダイレクト等）で期待と異なる場合があるが、`--output text` で明示上書き可能

## Verification / Guardrails

- `go build ./...` および `go test ./...` がパスすること
- `rgw <cmd> --output json` で有効なJSONが出力されること
- `rgw <cmd> | cat` （非TTY）でJSONがデフォルト出力されること
- `rgw <cmd> --output text | cat` でテキスト出力が強制されること
