# ADR: rgw利用ガイドをClaude Codeスキルとして提供する

- Status: Proposed
- Date: 2026-03-31

## Context

rgw はAIエージェントが外部ツールとして利用するCLIだが、`--help` や `rgw describe` だけでは伝わらないワークフロー（「リンク前にstatusを確認する」等）やルール（「常に --output json を使う」等）がある。

以前は CONTEXT.md に記載していたが、ランタイムで取得可能な情報と重複しており、AGENTS.md（開発者向け）とは対象読者が異なるため削除した。

ユーザー向けのワークフローガイドを提供する手段が必要になった。

## Decision

`.claude/skills/rgw-guide/SKILL.md` にClaude Codeプロジェクトスキルとして作成する。

## Decision Details

- rgw の使い方、ワークスペース操作、symlink管理について聞かれたときにトリガーされる
- ディレクトリ配置規約、典型ワークフロー、エージェント向けルール、コマンドリファレンス、トラブルシューティングを含む
- トークン効率のため英語で記述する
- 将来的に `rgw guide` サブコマンドやClaude Codeプラグインへの移行が可能

## Alternatives Considered

- **CONTEXT.md を維持**: ランタイムイントロスペクションと重複し、AGENTS.mdとは対象読者が異なる
- **AGENTS.md に追記**: 開発者向けガイドにユーザー向け内容が混在し、actionable ratioが低下する
- **Claude Codeプラグイン**: プロジェクト固有スキルに対して配布オーバーヘッドが過大。需要が出てから検討する
- **`rgw guide` サブコマンドを即実装**: 実装コストが高い。まずスキルとして内容を検証し、必要に応じてバイナリに組み込む

## Consequences

- このリポジトリでClaude Codeを使うユーザーはスキルトリガーでワークフローガイドを取得できる
- マークダウンファイルなので内容の反復改善が容易
- `go install` でバイナリだけ入れたユーザーにはスキルが届かない。現時点では許容する
- `rgw guide` サブコマンドやプラグインへの将来的な移行パスは残る

## Verification / Guardrails

- `.claude/skills/rgw-guide/SKILL.md` が存在しスキルのfrontmatter形式に従っていること
- rgw の使い方を聞かれたときにスキルが正しくトリガーされること
- AGENTS.md にユーザー向け内容が混入していないこと
