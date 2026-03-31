---
name: rgw-guide
description: Usage guide for rgw CLI — manages symlinks between Git worktrees and ROS workspace src/ directories. Use this skill when the user asks how to use rgw, manage workspaces, link worktrees, switch branches, diagnose issues, or anything related to rgw commands and workflows. Also trigger when someone mentions symlink management for ROS workspaces, ghq + worktree workflows, or colcon workspace setup.
triggers:
  - rgw usage
  - rgw workflow
  - workspace management
  - symlink management
  - worktree switching
  - colcon workspace
---

# rgw Usage Guide

rgw manages symlinks between Git worktrees and ROS workspace `src/` directories.
It does **not** create worktrees, run Git operations, or perform builds.

**Prerequisites:** Worktrees must already exist (created via `git worktree add` or gwq). rgw only manages symlinks to them.

## Directory Layout

rgw assumes this structure (ghq + gwq convention):

```
~/ghq/github.com/org/repo/              # main repo (ghq-managed)
~/worktree/github.com/org/repo/         # worktrees (one per branch)
  feature-branch-a/
  fix-something/
~/ros2_ws/src/                           # ROS workspace — rgw manages symlinks here
  repo -> ~/worktree/.../feature-branch-a/
```

## Common Workflows

### Initial setup
```bash
rgw ws add --name ros2 --path ~/ros2_ws
```

### Link a repo to the workspace
```bash
rgw wt list org/repo                         # check available worktrees
rgw link set org/repo --branch feature/x     # create symlink
rgw link status                              # verify
```

### Switch branches (quick)
`switch` takes the **alias name** (as shown in `rgw link status`), not org/repo:
```bash
rgw switch robot_nav feature/other-branch    # alias + branch
rgw switch robot_nav -i                      # interactive picker
```

### Manage workspaces
```bash
rgw ws list                                  # list all workspaces
rgw ws use nav                               # switch default workspace
rgw ws current                               # show active workspace
```

### Diagnose and repair
```bash
rgw doctor                                   # environment health check
rgw link repair                              # remove broken symlinks
```

## Rules for AI Agents

These rules are not obvious from `--help` and matter for safe automation:

1. **Always use `--output json`** — non-TTY defaults to JSON automatically, but be explicit
2. **Always use `--dry-run` before mutating** — `link set`, `link unset`, `link repair`, `ws add`, `ws use`
3. **Never guess repo names** — run `rgw wt list` or `rgw link status --output json` first
4. **Check before linking** — `rgw link status` shows current state; don't blindly overwrite
5. **Use `rgw describe <command>`** to query command schema at runtime instead of hardcoding assumptions

## Agent Workflow (recommended sequence)

```
rgw ws list --output json                    # 1. discover workspaces
rgw wt list <repo> --output json             # 2. list available branches
rgw link status --output json                # 3. check current state
rgw link set <repo> --branch <b> --dry-run   # 4. preview change
rgw link set <repo> --branch <b>             # 5. execute
rgw link status --output json                # 6. confirm result
```

## Repo Argument Formats

| Format | Example | Notes |
|--------|---------|-------|
| `repo` | `robot_nav` | Searches ghq root; must be unambiguous |
| `org/repo` | `your-org/robot_nav` | Assumes github.com |
| `host/org/repo` | `github.com/your-org/robot_nav` | Fully qualified |

## Global Flags

| Flag | Purpose |
|------|---------|
| `-o, --output json` | JSON output (default in non-TTY) |
| `--dry-run` | Preview changes without executing |
| `-w, --ws <name>` | Override active workspace |

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `workspace "X" not found` | Typo or not registered | Run `rgw ws list` to see valid names |
| `repository "X" not found` | Short name ambiguous or missing | Run `rgw wt list org/repo` with full org/repo format |
| `branch "X" not found` | Branch has no worktree | The error output lists available branches. Create the worktree first with `git worktree add`, then retry |
| `alias "X" not found` | Alias not linked yet | Run `rgw link status` to see existing aliases; use `rgw link set` to create one |
| `ambiguous repository` | Multiple repos with same name | Use `org/repo` to disambiguate |
| colcon `Duplicate package names` | Worktree inside ROS workspace | Move worktrees outside workspace; run `rgw link status` to check symlinks; use `COLCON_IGNORE` if needed |

## Key Constraint

rgw depends on ghq's directory layout (`host/org/repo`). Without ghq, set `RGW_GHQ_ROOT` to the directory containing your repos in that structure, or use `--path` for direct paths.
