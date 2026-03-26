# rgw 設計仕様書（Draft v1）

## 1. 概要

`rgw` は、Git worktree / jj workspace などの**既存の作業実体を検出し**、
ROS（特に colcon）ワークスペースの `src/` に対して **symlink を張り替えることで可変ビューを提供する CLI ツール**である。

本ツールは、`ghq`・`gwq`・`wtp`・`git-wt` 等のツールと**併用されることを前提**とし、
それらの機能を置き換えるものではない。

---

## 2. 設計方針

### 2.1 単一責務

rgw は以下に特化する：

* 作業実体（worktree / workspace）の検出
* ROS ws への露出（symlink）
* 状態の可視化
* 環境整合性の診断

### 2.2 ツール非依存

* worktree の作成・管理ツールには依存しない
* `git worktree` の状態を唯一の事実源とする
* jj は将来拡張として扱う

### 2.3 コピー禁止

* ファイルコピーは一切行わない
* 常に同一実体を参照する

---

## 3. スコープ

## 3.1 やること（Responsibilities）

### 3.1.1 Worktree の検出

* `git worktree list --porcelain` を用いて列挙
* 各 worktree に対して以下を取得：

  * path
  * branch
  * HEAD
  * dirty 状態

---

### 3.1.2 ROS ws への露出（symlink 管理）

* `<ros_ws>/src/<alias>` → `<worktree_path>` を作成
* `ln -sfn` による安全な置き換え
* 複数 ws をサポート

---

### 3.1.3 状態の可視化

* 現在のリンク状態を表示
* repo → worktree の対応関係を明示

---

### 3.1.4 安全性チェック

以下を検出・警告：

* build/install/log の存在
* symlink の破損
* worktree の dirty 状態
* コンテナ内での symlink 解決不可

---

### 3.1.5 ワークスペース管理

* 複数 ROS ws の定義・切替
* 環境変数および設定ファイルによる指定

---

### 3.1.6 インタラクティブ選択

* worktree 一覧から選択（fzf想定）
* 表示項目：

  * alias
  * branch
  * dirty
  * path

---

### 3.1.7 開発支援

* 実体パスを開く（VS Code 等）
* 差分確認補助（任意）

---

## 3.2 やらないこと（Non-Goals）

### 3.2.1 Git 操作

* checkout / switch
* commit
* push / pull
* merge / rebase

---

### 3.2.2 Worktree 管理

* worktree の作成・削除（初期スコープ外）
* ブランチ作成

---

### 3.2.3 ビルド操作

* colcon build の実行
* ROS 環境セットアップ

---

### 3.2.4 VCS 依存の抽象化

* Git / jj の完全統一インターフェース（将来対応）

---

## 4. CLI 仕様

## 4.1 ワークスペース管理

```
rgw ws list
rgw ws add --name <name> --path <path>
rgw ws use <name>
rgw ws current
```

---

## 4.2 Worktree 検出

```
rgw wt list <repo>
```

出力：

* path
* branch
* dirty
* HEAD

---

## 4.3 リンク操作

### 設定

```
rgw link set <repo> --path <path> [-w <ws>]
rgw link set <repo> --branch <branch> [-w <ws>]
rgw link set <repo> --interactive [-w <ws>]
```

### 状態

```
rgw link status [-w <ws>] [--all-ws]
```

### 削除

```
rgw link unset <repo> [-w <ws>]
```

### 修復

```
rgw link repair [-w <ws>]
```

---

## 4.4 補助コマンド

```
rgw open <repo> [--path|--interactive]
rgw doctor
```

---

## 5. 設定仕様

### 設定ファイル

```
~/.config/rgw/config.toml
```

### 例

```toml
[ghq]
root = "~/ghq"

[ros]
[[ros.workspaces]]
name = "default"
path = "~/ros2_ws"
src_subdir = "src"

[defaults]
ros_workspace = "default"

[alias]
mode = "org_repo"
```

---

### 環境変数

| 変数             | 内容        |
| -------------- | --------- |
| RGW_WS         | 使用する ws 名 |
| RGW_WS_PATH    | ws パス直接指定 |
| RGW_GHQ_ROOT   | ghq root  |
| RGW_ALIAS_MODE | alias モード |

---

## 6. Alias ルール

衝突回避のため複数モードを用意：

| mode          | 例                     |
| ------------- | --------------------- |
| repo          | repo                  |
| org_repo      | org__repo             |
| host_org_repo | github.com__org__repo |

デフォルト：`org_repo`

---

## 7. データ管理

### 方針

* worktree 情報は保持しない（常に動的取得）
* 永続化するのは以下のみ：

#### 必須

* ws 定義

#### 任意（将来）

* link 状態
* alias マッピング

---

## 8. 安全設計

### 8.1 ws 汚染チェック

* build/install/log 存在時に警告または拒否

---

### 8.2 worktree 状態チェック

* dirty 時に warn/deny

---

### 8.3 symlink 整合性

* 存在しないパス検出
* コンテナ内解決確認

---

### 8.4 編集場所の明確化

* `rgw open` は常に実体側を開く
* symlink 側での編集を推奨しない

---

## 9. 典型ワークフロー

### 9.1 作業開始

1. 任意ツールで worktree 作成
2. `rgw wt list repo`
3. `rgw link set repo --interactive`

---

### 9.2 ビルド

```
cd ros_ws
colcon build
```

---

### 9.3 切替

```
rgw link set repo --interactive
```

---

### 9.4 状態確認

```
rgw link status
```

---

## 10. 拡張可能性（将来）

* jj workspace 対応
* state DB（SQLite）
* PR 補助
* LSP / build cache 最適化
* コンテナ自動検出

---

## 11. 設計上の決定事項（重要）

* checkout は **行わない**
* worktree は **作らない**
* 真実は **git worktree**
* rgw は **ROS ws のビュー制御ツール**

---

## 12. まとめ

rgw は：

* ghq の「取得」
* gwq/wtp の「作業管理」

の上に乗る、

**「ROSワークスペースの可変ビュー制御レイヤ」**

として設計する。

---

