# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Build & Test Commands

```bash
make build              # Build binary to bin/rgw
make test               # Run all tests (go test ./...)
make lint               # Run golangci-lint
make install            # Install to $GOPATH/bin
go test -v ./internal/config/...   # Run tests for a single package
go test -run TestParsePorcelain ./internal/worktree/...  # Run a single test
```

## Architecture

rgw is a Go CLI tool that manages symlinks between Git worktrees (managed by ghq) and ROS workspace `src/` directories. It never performs Git operations, creates worktrees, or copies files -- it only creates/removes symlinks.

### Package dependency flow

```
cmd/rgw/main.go → internal/cli (Cobra commands)
                     ├── config     ← TOML config + env vars + workspace resolution
                     ├── ghq        ← Parse repo arguments, resolve to filesystem paths
                     ├── worktree   ← Parse `git worktree list --porcelain` output
                     ├── alias      ← Symlink naming: repo / org_repo / host_org_repo
                     ├── symlink    ← Create/remove symlinks in workspace src/
                     ├── workspace  ← Workspace CRUD (add/use/list/current)
                     ├── validate   ← Security boundary: path traversal, control chars
                     ├── selector   ← Bubbletea interactive worktree picker
                     ├── doctor     ← Environment health checks
                     └── output     ← JSON/text output with TTY auto-detection
```

### Key design decisions (documented in docs/adr/)

- **ADR-0001**: Non-TTY output defaults to JSON; TTY defaults to text. Controlled by `--output` flag and `mattn/go-isatty`.
- **ADR-0002**: All external input is validated in the `validate` package at CLI boundaries. Internal/Git outputs are trusted.
- **ADR-0003**: `--dry-run` flag on mutating commands and `describe` command for runtime introspection, designed for AI agent safety.

### Input validation model

The `validate` package is the security boundary. All user-supplied paths, repo names, and workspace names must pass through it before use. `SafePath()` prevents path traversal; `RepoSegment()` rejects `..` and special characters. Symlink operations refuse to remove non-symlinks.

## Conventions

- Config file: `~/.config/rgw/config.toml` (respects `XDG_CONFIG_HOME`)
- Workspace resolution precedence: `--ws` flag > `RGW_WS` env > `RGW_WS_PATH` env > config default > first workspace
- Repo argument formats: `repo`, `org/repo`, `host/org/repo`
- ADRs go in `docs/adr/` using the template below
- Design spec is in Japanese (`DESIGN.md`); agent-facing guidance is in this file, and the legacy `CONTEXT.md` entrypoint redirects to `.claude/skills/rgw-guide/SKILL.md`

## ADRs

Use an Architecture Decision Record when a change needs to preserve the reason behind a design choice, not just the implementation details.
When reviewing changes, explicitly check whether the PR introduces or materially changes a design or policy decision that should be captured as an ADR, and flag it if the rationale is not documented.

- Store ADRs in `docs/adr/`
- Use file names like `NNNN-short-kebab-case.md`
- Prefer `Proposed` while implementation is in progress, then update to `Accepted` once shipped
- Keep the ADR focused on one decision

### ADR Template

```md
# ADR: <short decision title>

- Status: Proposed
- Date: YYYY-MM-DD

## Context

Describe the technical background, current constraints, and the specific problem that requires a decision.

## Decision

State the decision clearly and directly.

## Decision Details

Document the concrete rules, scope, and implementation-facing behavior that follow from the decision.

## Alternatives Considered

List the main alternatives and why they were not chosen.

## Consequences

Describe the positive effects, negative effects, and ongoing maintenance cost.

## Verification / Guardrails

Capture the invariants, tests, and checks that should hold after the decision is implemented.
```
