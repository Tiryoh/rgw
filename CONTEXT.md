# rgw - AI Agent Context

## Overview

rgw is a CLI tool that manages symlinks between Git worktrees (managed by ghq) and ROS workspace `src/` directories. It does **not** create worktrees, run Git operations, or perform builds.

## Commands

| Command | Description | Args |
|---------|-------------|------|
| `ws list` | List configured workspaces | |
| `ws add` | Add a new workspace | `--name`, `--path` (required) |
| `ws use <name>` | Set default workspace | name |
| `ws current` | Show active workspace | |
| `wt list <repo>` | List worktrees for a repository | repo |
| `link set <repo>` | Create/update symlink | repo, `--branch`, `--path`, `-i` |
| `link unset <repo>` | Remove symlink | repo |
| `link status` | Show symlink status | `--all-ws` |
| `link repair` | Remove broken symlinks | |
| `doctor` | Check environment health | |
| `open <repo>` | Open worktree in editor | repo, `--path`, `-i` |
| `describe [cmd...]` | Machine-readable command schema | command path |

## Global Flags

| Flag | Description |
|------|-------------|
| `-w, --ws <name>` | Override active workspace |
| `-o, --output <format>` | Output format: `text` or `json` (auto-detects TTY) |
| `--dry-run` | Show what would happen without making changes |
| `-v, --verbose` | Verbose output |

## Repo Argument Formats

- `repo` - searches ghq root (must be unambiguous)
- `org/repo` - assumes github.com
- `host/org/repo` - fully qualified

## Agent Workflow

Always follow this sequence for symlink operations:

1. **Discover**: `rgw ws list --output json` to find available workspaces
2. **Verify**: `rgw wt list <repo> --output json` to list available worktrees/branches
3. **Check**: `rgw link status --output json` to understand current state
4. **Preview**: `rgw link set <repo> --branch <branch> --dry-run --output json` to verify the action
5. **Execute**: `rgw link set <repo> --branch <branch> --output json` to perform the change
6. **Confirm**: `rgw link status --output json` to verify the result

## Rules for AI Agents

- **Always use `--output json`** for machine-readable output
- **Always use `--dry-run` before mutating operations** (`link set`, `link unset`, `link repair`, `ws add`, `ws use`)
- **Never guess repository names** - always list and verify with `wt list` first
- **Always check `link status` before `link set`** to understand current symlink state
- **Use `wt list <repo>` to verify branch names** before using `--branch`
- **Use `doctor --output json`** to diagnose environment issues
- **Use `describe <command>`** to get machine-readable command metadata at runtime

## Environment Variables

| Variable | Description |
|----------|-------------|
| `RGW_CONFIG` | Override config file path |
| `RGW_WS` | Override active workspace name |
| `RGW_WS_PATH` | Override active workspace path |
| `RGW_GHQ_ROOT` | Override ghq root directory |
| `RGW_ALIAS_MODE` | Alias mode: `repo`, `org_repo` (default), `host_org_repo` |
| `EDITOR` | Editor for `open` command (default: `code`) |

## Configuration

Config file: `~/.config/rgw/config.toml` (respects `XDG_CONFIG_HOME`)

```toml
[ghq]
root = "~/.ghq"

[alias]
mode = "org_repo"

[[ros.workspaces]]
name = "my_ws"
path = "~/ros2_ws"
src_subdir = "src"

[defaults]
ros_workspace = "my_ws"
```

## Error Handling

- Errors are written to stderr; structured output goes to stdout
- Non-zero exit code indicates failure
- Common errors:
  - `workspace "X" not found` - use `ws list` to find valid names
  - `repository "X" not found` - use full `org/repo` or `host/org/repo` format
  - `branch "X" not found` - use `wt list <repo>` to see available branches
  - `ambiguous repository` - specify as `org/repo` to disambiguate
