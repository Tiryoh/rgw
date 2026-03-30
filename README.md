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

```bash
rgw ws add --name default --path ~/ros2_ws           # ワークスペース登録
rgw wt list your-org/robot_nav                        # worktree 確認
rgw link set your-org/robot_nav --branch feature/x    # リンク作成
rgw link status                                       # 状態確認
rgw link set your-org/robot_nav --branch main         # 切替
```

詳しくは後述の[チュートリアル](#チュートリアル)を参照してください。

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
mode = "repo"
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
| `repo` (デフォルト) | `my_pkg` |
| `org_repo` | `Tiryoh__my_pkg` |
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

## チュートリアル

ghq + gwq (git worktree) + rgw を組み合わせた典型的なワークフローを紹介します。

### 前提: ディレクトリ構成

```
~/ghq/                                          # ghq root（メインリポジトリ）
  github.com/
    your-org/
      robot_nav/                                # メインブランチのチェックアウト
      robot_sensors/

~/worktree/                                     # gwq root（worktree 群）
  github.com/
    your-org/
      robot_nav/
        feature-add-obstacle-detection/         # worktree（ブランチごと）
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

### Step 2: worktree を作成する（rgw のスコープ外）

gwq や `git worktree add` で worktree を作成します。rgw は worktree の作成・削除を行いません。

```bash
cd ~/ghq/github.com/your-org/robot_nav
git worktree add ~/worktree/github.com/your-org/robot_nav/feature-add-obstacle-detection feature/add-obstacle-detection
```

> **注意: worktree の配置場所について**
>
> worktree は必ず **ROS ワークスペースの外**に作成してください。リポジトリ内（例: `.claude/` や `.worktrees/`）に worktree を作ると、colcon がパッケージを二重に検出して `Duplicate package names not supported` エラーになります。
>
> rgw が symlink 経由でリンクしたリポジトリ内に `package.xml` を含むサブディレクトリがある場合も同様です。問題が発生した場合は、該当ディレクトリに空の `COLCON_IGNORE` ファイルを置くことで colcon の探索対象から除外できます。

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

これで `~/ros2_ws/src/robot_nav` が worktree への symlink になります。

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

別のブランチに切り替えたいときは、同じ `link set` で上書きします。対話モードで選ぶこともできます。

```bash
# 対話選択で切替
rgw link set your-org/robot_nav --interactive

# または直接指定
rgw link set your-org/robot_nav --branch main
```

symlink の向き先が変わるだけなので、切替は一瞬です。

### Step 7: 環境を診断する

```bash
rgw doctor
```

```
[OK  ]  ghq_root: ghq root: /home/user/ghq
[OK  ]  ros2/path: Workspace: /home/user/ros2_ws
[WARN]  ros2/build: build/ directory exists (stale build artifacts?)
[OK  ]  ros2/symlinks: No broken symlinks
```

`build/` が残っている場合は `colcon build` のキャッシュが古いブランチのものかもしれません。必要に応じて `rm -rf build/ install/ log/` してからリビルドしてください。

### Step 8: リンクを解除する

```bash
rgw link unset your-org/robot_nav
```

```
Unlinked robot_nav
```

## 典型ワークフロー（まとめ）

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

- **symlink のみ** — ファイルコピーは一切行わない
- **Git 操作しない** — checkout, commit, push 等は行わない
- **Worktree を作らない** — 作成・削除は別ツールの責務
- **真実は git worktree** — `git worktree list` の出力を唯一の事実源とする

詳細は [DESIGN.md](DESIGN.md) を参照。

## ライセンス

TBD
