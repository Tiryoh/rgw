# Agent-Friendly CLI Design

- Source: Justin Poehnelt (Google Senior DevRel), 2026-03-04
- URL: https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/

## Core Message

- Human DX = discoverability and forgiveness
- Agent DX = predictability and defense in depth
- These are orthogonal concerns. Retrofitting a human CLI for agents is the wrong approach.
- "Agents are not trusted operators" — the guiding design philosophy.

## 7 Principles

### 1. Raw JSON Payloads > Individual Flags

- LLMs excel at JSON generation — zero translation loss
- Support both human-friendly flags and `--json` / `--output json`
- Default to NDJSON when stdout is not a TTY

### 2. Runtime Schema Introspection

- CLI itself should return schemas (machine-readable JSON)
- CLI becomes the canonical source of "what the current API accepts"

### 3. Context Window Conservation

- Field masks: fetch only needed fields
- NDJSON pagination: stream objects without buffering top-level arrays

### 4. Input Validation Against Hallucinations

- CLI is the last line of defense
- Key threats and mitigations:

| Threat | Agent Mistake | Mitigation |
|--------|--------------|------------|
| File paths | Path traversal (`../../.ssh`) | Sandbox to CWD |
| Control chars | Invisible characters | Reject ASCII < 0x20 |
| Resource IDs | Query params in IDs (`fileId?fields=name`) | Reject `?` `#` |
| URL encoding | Pre-encoding → double-encoding (`%2e%2e`) | Reject `%` |

### 5. Skill Files

- `SKILL.md` / `CONTEXT.md`: YAML frontmatter + structured Markdown
- Document agent-specific invariants not obvious from `--help`

### 6. Multi-Surface Support

- Same binary serves CLI, MCP (stdio JSON-RPC), extensions, env var auth
- Single source of truth (e.g., Discovery Document) drives all surfaces

### 7. Safety Mechanisms: dry-run + Response Sanitization

- `--dry-run`: validate locally without calling APIs
- `--sanitize`: sanitize API responses before returning to agents (prompt injection defense)

## Recommended Retrofit Order

1. `--output json`
2. Input validation
3. Schema / `--describe`
4. Field masks / `--fields`
5. `--dry-run`
6. `CONTEXT.md` / skill files
7. MCP support
