# rgw

Git worktree と ROS ワークスペースを symlink でつなぐ CLI ツールです。

`ghq` や `git worktree` で管理しているリポジトリを、ROS（colcon）ワークスペースの `src/` 配下に symlink として配置できます。ブランチを切り替えるときは symlink のリンク先を差し替えるだけなので、ファイルのコピーやチェックアウトのやり直しは不要です。

## 特徴

- **symlink だけで運用**: ファイルをコピーせず、常に元の作業ツリーを直接参照します
- **worktree を自動検出**: `git worktree list` の出力から、現在の状態を取得します
- **複数ワークスペースに対応**: プロジェクトごとに ROS ワークスペースを使い分けられます
- **対話的に選択可能**: fuzzy finder で worktree を選び、そのままリンクできます
- **環境を診断**: `rgw doctor` でリンク切れや古いビルド成果物をまとめて確認できます
- **シェル補完に対応**: bash / zsh / fish で、リポジトリ名やワークスペース名も補完できます

## インストール

```bash
go install github.com/Tiryoh/rgw/cmd/rgw@latest
```

ソースからビルドする場合:

```bash
git clone https://github.com/Tiryoh/rgw.git
cd rgw
make build    # bin/rgw を生成
make install  # $GOPATH/bin にインストール
```

## クイックスタート

```bash
rgw ws add --name default --path ~/ros2_ws           # ワークスペースを登録
rgw wt list your-org/robot_nav                       # worktree を確認
rgw link set your-org/robot_nav --branch feature/x   # リンクを作成
rgw link status                                      # 状態を確認
rgw link set your-org/robot_nav --branch main        # 別のブランチへ切り替え
```

詳しい使い方は [チュートリアル](#チュートリアル) を参照してください。

## コマンド一覧

### ワークスペース管理 (`rgw ws`)

| コマンド | 説明 |
|---|---|
| `rgw ws list` | 登録済みワークスペースの一覧を表示 |
| `rgw ws add --name <name> --path <path>` | ワークスペースを追加 |
| `rgw ws use <name>` | デフォルトのワークスペースを設定 |
| `rgw ws current` | 現在使用中のワークスペースを表示 |

### worktree 一覧 (`rgw wt`)

| コマンド | 説明 |
|---|---|
| `rgw wt list <repo>` | 指定したリポジトリの worktree を一覧表示 |

### リンク操作 (`rgw link`)

| コマンド | 説明 |
|---|---|
| `rgw link set <repo> --branch <branch>` | ブランチを指定してリンク |
| `rgw link set <repo> --path <path>` | パスを指定してリンク |
| `rgw link set <repo> --interactive` | 対話的に選択してリンク |
| `rgw link status [--all-ws]` | リンクの状態を表示 |
| `rgw link unset <repo>` | リンクを削除 |
| `rgw link repair` | 壊れたリンクを修復 |

### その他

| コマンド | 説明 |
|---|---|
| `rgw open <repo> [--interactive]` | worktree をエディタで開く |
| `rgw doctor` | 環境の状態を確認 |

### グローバルフラグ

| フラグ | 説明 |
|---|---|
| `-w, --ws <name>` | 使用するワークスペースを一時的に切り替える |
| `-v, --verbose` | 詳細な情報を表示 |

## 設定

### 設定ファイル

設定は `~/.config/rgw/config.toml` に保存されます（`XDG_CONFIG_HOME` が設定されている場合はそちらを優先します）。

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
mode = "repo"
```

### 環境変数

| 変数 | 説明 | 例 |
|---|---|---|
| `RGW_WS` | 使用するワークスペース名 | `default` |
| `RGW_WS_PATH` | 使用するワークスペースのパスを直接指定 | `~/ros2_ws` |
| `RGW_GHQ_ROOT` | ghq のルートディレクトリ | `~/ghq` |
| `RGW_ALIAS_MODE` | alias モード | `org_repo` |

設定の優先順位: **コマンドラインフラグ > 環境変数 > config.toml > デフォルト値**

### alias モード

ワークスペースに作成する symlink 名の付け方を選べます。

| モード | 例 (`github.com/Tiryoh/my_pkg`) |
|---|---|
| `repo` (デフォルト) | `my_pkg` |
| `org_repo` | `Tiryoh__my_pkg` |
| `host_org_repo` | `github.com__Tiryoh__my_pkg` |

## シェル補完

### Bash

```bash
# 現在のシェルで有効化
source <(rgw completion bash)

# 永続化
rgw completion bash > /etc/bash_completion.d/rgw
```

### Zsh

```zsh
# 現在のシェルで有効化
source <(rgw completion zsh)

# 永続化（oh-my-zsh の場合）
rgw completion zsh > "${fpath[1]}/_rgw"
```

### Fish

```fish
# 現在のシェルで有効化
rgw completion fish | source

# 永続化
rgw completion fish > ~/.config/fish/completions/rgw.fish
```

サブコマンドやフラグだけでなく、ghq 配下のリポジトリ名やワークスペース名も補完されます。

## チュートリアル

ghq、git worktree、rgw を組み合わせた基本的な使い方を紹介します。

### 前提: ディレクトリ構成

```
~/ghq/                                          # ghq ルート（メインリポジトリ）
  github.com/
    your-org/
      robot_nav/                                # メインブランチのチェックアウト
      robot_sensors/

~/worktree/                                     # worktree 用ディレクトリ
  github.com/
    your-org/
      robot_nav/
        feature-add-obstacle-detection/         # ブランチごとの worktree
        feature-update-costmap/
        fix-planner-timeout/
      robot_sensors/
        feature-lidar-filter/

~/ros2_ws/                                      # ROS ワークスペース
  src/                                          # ← rgw がここに symlink を管理
```

### Step 1: ワークスペースを登録する

```bash
rgw ws add --name ros2 --path ~/ros2_ws
```

### Step 2: worktree を作成する（rgw の対象外）

worktree の作成は `git worktree add` や gwq などで行います。rgw は worktree の作成や削除は行わず、既存の worktree を検出して利用します。

```bash
cd ~/ghq/github.com/your-org/robot_nav
git worktree add ~/worktree/github.com/your-org/robot_nav/feature-add-obstacle-detection feature/add-obstacle-detection
```

> **worktree の配置場所に注意**
>
> worktree は **ROS ワークスペースの外側**に作成してください。ワークスペース内に置くと、colcon がパッケージを二重に検出し、`Duplicate package names not supported` エラーになることがあります。
>
> symlink 先のリポジトリに `package.xml` を含むサブディレクトリがある場合も同様です。問題が起きる場合は、そのディレクトリに空の `COLCON_IGNORE` を置くことで、colcon の検出対象から外せます。

### Step 3: 利用可能な worktree を確認する

```bash
rgw wt list your-org/robot_nav
```

```
BRANCH                          HEAD      PATH
main                            a1b2c3d4  /home/user/ghq/github.com/your-org/robot_nav
feature/add-obstacle-detection  e5f6g7h8  /home/user/worktree/github.com/your-org/robot_nav/feature-add-obstacle-detection
feature/update-costmap          i9j0k1l2  /home/user/worktree/github.com/your-org/robot_nav/feature-update-costmap
fix/planner-timeout             m3n4o5p6  /home/user/worktree/github.com/your-org/robot_nav/fix-planner-timeout
```

### Step 4: ワークスペースにリンクする

```bash
# ブランチ名で指定
rgw link set your-org/robot_nav --branch feature/add-obstacle-detection
```

```
Linked robot_nav -> /home/user/worktree/github.com/your-org/robot_nav/feature-add-obstacle-detection
```

これで `~/ros2_ws/src/robot_nav` は、指定した worktree を指す symlink になります。

### Step 5: 複数パッケージをリンクしてビルドする

```bash
# センサーパッケージもリンク
rgw link set your-org/robot_sensors --branch feature/lidar-filter

# リンク状態を確認
rgw link status
```

```
ALIAS            TARGET                                                                                  STATUS
robot_nav        /home/user/worktree/github.com/your-org/robot_nav/feature-add-obstacle-detection        ok
robot_sensors    /home/user/worktree/github.com/your-org/robot_sensors/feature-lidar-filter              ok
```

```bash
# ビルド
cd ~/ros2_ws && colcon build
```

### Step 6: ブランチを切り替える

同じ `link set` を実行するだけで、symlink の向き先が切り替わります。対話モードで選ぶこともできます。

```bash
# 対話選択で切替
rgw link set your-org/robot_nav --interactive

# または直接指定
rgw link set your-org/robot_nav --branch main
```

symlink を張り替えるだけなので一瞬です。

### Step 7: 環境をチェックする

```bash
rgw doctor
```

```
[OK  ]  ghq_root: ghq root: /home/user/ghq
[OK  ]  ros2/path: Workspace: /home/user/ros2_ws
[WARN]  ros2/build: build/ directory exists (stale build artifacts?)
[OK  ]  ros2/symlinks: No broken symlinks
```

`build/` が残っていると警告が出ます。古いブランチのビルド成果物かもしれないので、必要に応じて `rm -rf build/ install/ log/` してからリビルドしてください。

### Step 8: リンクを解除する

```bash
rgw link unset your-org/robot_nav
```

```
Unlinked robot_nav
```

## まとめ: 一連の流れ

```bash
# worktree 作成 → 確認 → リンク → ビルド → 切替 → 診断
git worktree add ~/worktree/github.com/your-org/robot_nav/feature-x feature/x
rgw wt list your-org/robot_nav
rgw link set your-org/robot_nav --branch feature/x
cd ~/ros2_ws && colcon build
rgw link set your-org/robot_nav --branch main    # 切替
rgw doctor                                        # 診断
```

## 設計方針

- **symlink だけ** — ファイルのコピーは一切しない
- **Git には触らない** — checkout, commit, push などは行わない
- **worktree は作らない** — 作成・削除は git や ghq の仕事
- **情報源は git worktree** — `git worktree list` の出力だけを信頼する

詳しくは [DESIGN.md](DESIGN.md) を参照してください。

## ライセンス

TBD
