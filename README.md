# rgw

ROS ワークスペースのビューコントローラ for Git worktrees.

`rgw` は Git worktree を検出し、ROS (colcon) ワークスペースの `src/` ディレクトリに symlink を作成・管理する CLI ツールです。`ghq` や `git worktree` と併用し、ブランチごとの worktree を ROS ワークスペースに自在に切り替えられます。

## 特徴

- **symlink ベース** — ファイルコピーなし。常に同一実体を参照
- **動的検出** — `git worktree list --porcelain` から常に最新の worktree 情報を取得
- **複数ワークスペース対応** — ROS ワークスペースを複数管理・切替
- **対話選択** — bubbletea ベースの fuzzy finder で worktree を選択
- **環境診断** — `rgw doctor` でリンク切れ、build 残骸、dirty worktree を検出
- **シェル補完** — bash / zsh / fish 対応。リポジトリ名・ワークスペース名を動的補完

## インストール

```bash
go install github.com/Tiryoh/rgw/cmd/rgw@latest
```

ソースからビルド:

```bash
git clone https://github.com/Tiryoh/rgw.git
cd rgw
make build    # bin/rgw に出力
make install  # $GOPATH/bin にインストール
```

## クイックスタート

### 1. ワークスペースを登録

```bash
rgw ws add --name default --path ~/ros2_ws
```

### 2. リポジトリの worktree を確認

```bash
rgw wt list Tiryoh/my_robot
```

### 3. symlink を作成

```bash
# ブランチ指定
rgw link set Tiryoh/my_robot --branch feature/nav

# 対話選択
rgw link set Tiryoh/my_robot --interactive
```

### 4. 状態確認

```bash
rgw link status
```

```
ALIAS                  TARGET                                          STATUS
Tiryoh__my_robot       /home/user/ghq/github.com/Tiryoh/my_robot-nav  ok
```

### 5. 切り替え

```bash
rgw link set Tiryoh/my_robot --branch main
```

## コマンド一覧

### ワークスペース管理 (`rgw ws`)

| コマンド | 説明 |
|---|---|
| `rgw ws list` | 登録済みワークスペース一覧 |
| `rgw ws add --name <name> --path <path>` | ワークスペース追加 |
| `rgw ws use <name>` | デフォルトワークスペース設定 |
| `rgw ws current` | 現在のアクティブワークスペース表示 |

### Worktree 検出 (`rgw wt`)

| コマンド | 説明 |
|---|---|
| `rgw wt list <repo>` | リポジトリの worktree 一覧 |

### リンク操作 (`rgw link`)

| コマンド | 説明 |
|---|---|
| `rgw link set <repo> --branch <branch>` | ブランチ指定でリンク作成 |
| `rgw link set <repo> --path <path>` | パス指定でリンク作成 |
| `rgw link set <repo> --interactive` | 対話選択でリンク作成 |
| `rgw link status [--all-ws]` | リンク状態表示 |
| `rgw link unset <repo>` | リンク削除 |
| `rgw link repair` | 壊れたリンクを修復 |

### 補助コマンド

| コマンド | 説明 |
|---|---|
| `rgw open <repo> [--interactive]` | エディタで worktree を開く |
| `rgw doctor` | 環境診断 |

### グローバルフラグ

| フラグ | 説明 |
|---|---|
| `-w, --ws <name>` | 使用するワークスペースを一時的に指定 |
| `-v, --verbose` | 詳細出力 |

## 設定

### 設定ファイル

`~/.config/rgw/config.toml` (`XDG_CONFIG_HOME` を尊重):

```toml
[ghq]
root = "~/ghq"

[ros]
[[ros.workspaces]]
name = "default"
path = "~/ros2_ws"
src_subdir = "src"

[[ros.workspaces]]
name = "nav"
path = "~/nav_ws"

[defaults]
ros_workspace = "default"

[alias]
mode = "org_repo"
```

### 環境変数

| 変数 | 説明 | 例 |
|---|---|---|
| `RGW_WS` | 使用する ws 名 | `default` |
| `RGW_WS_PATH` | ws パス直接指定 | `~/ros2_ws` |
| `RGW_GHQ_ROOT` | ghq ルート | `~/ghq` |
| `RGW_ALIAS_MODE` | alias モード | `org_repo` |

設定の優先順位: **フラグ > 環境変数 > config.toml > デフォルト値**

### Alias モード

symlink 名の命名規則を制御します:

| モード | 例 (`github.com/Tiryoh/my_pkg`) |
|---|---|
| `repo` | `my_pkg` |
| `org_repo` (デフォルト) | `Tiryoh__my_pkg` |
| `host_org_repo` | `github.com__Tiryoh__my_pkg` |

## シェル補完

### Bash

```bash
# 現在のセッション
source <(rgw completion bash)

# 永続化
rgw completion bash > /etc/bash_completion.d/rgw
```

### Zsh

```zsh
# 現在のセッション
source <(rgw completion zsh)

# 永続化 (oh-my-zsh の場合)
rgw completion zsh > "${fpath[1]}/_rgw"
```

### Fish

```fish
rgw completion fish | source

# 永続化
rgw completion fish > ~/.config/fish/completions/rgw.fish
```

補完はサブコマンド、フラグに加え、`<repo>` 引数 (ghq 管理下のリポジトリ) とワークスペース名も動的に補完します。

## 典型ワークフロー

```bash
# 1. worktree を作成 (rgw のスコープ外)
cd ~/ghq/github.com/Tiryoh/my_robot
git worktree add ../my_robot-feature feature/nav

# 2. ROS ws にリンク
rgw link set Tiryoh/my_robot --branch feature/nav

# 3. ビルド
cd ~/ros2_ws && colcon build

# 4. 別ブランチに切替
rgw link set Tiryoh/my_robot --interactive

# 5. 状態確認
rgw link status

# 6. 環境診断
rgw doctor
```

## 設計方針

- **symlink のみ** — ファイルコピーは一切行わない
- **Git 操作しない** — checkout, commit, push 等は行わない
- **Worktree を作らない** — 作成・削除は別ツールの責務
- **真実は git worktree** — `git worktree list` の出力を唯一の事実源とする

詳細は [DESIGN.md](DESIGN.md) を参照。

## ライセンス

TBD
