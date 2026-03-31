---
name: agent-friendly-cli
description: Guidelines for designing and reviewing CLIs that AI agents use as primary consumers. Use this skill whenever working on CLI design, agent-facing interfaces, MCP-compatible CLIs, or reviewing a CLI for agent compatibility. Also trigger when someone asks about making a CLI "agent-friendly" or "machine-readable", or when designing --output json, --dry-run, input validation, or schema introspection features.
triggers:
  - CLI design
  - agent-friendly CLI
  - AI agent interface design
  - MCP-compatible CLI
  - machine-readable output
---

# Agent-Friendly CLI Design

Source: Justin Poehnelt (Google Senior DevRel), 2026-03-04
https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/

For full details, examples, and implementation rationale, read `reference/agent-friendly-cli-design.md`.

## Core Philosophy

- Human DX = discoverability and forgiveness
- Agent DX = predictability and defense-in-depth
- These are orthogonal. Retrofitting a human CLI for agents is the wrong approach.
- **"The agent is not a trusted operator."** — the guiding design principle.

## 7 Principles

### 1. Structured Output (Raw JSON Payloads)
- Provide `--output json` flag
- Default to JSON when stdout is not a TTY
- LLMs excel at JSON generation — zero translation loss

### 2. Runtime Schema Introspection
- CLI itself should return schemas as machine-readable JSON
- Don't stuff static docs into system prompts — let agents query at runtime

### 3. Context Window Conservation
- Field masks: return only needed fields
- NDJSON pagination: stream objects without buffering

### 4. Input Validation (Hallucination Defense)
- CLI is the last line of defense

| Threat | Mitigation |
|--------|-----------|
| Path traversal (`../../.ssh`) | Sandbox to CWD |
| Control characters | Reject ASCII < 0x20 |
| Query params in resource IDs | Reject `?` `#` |
| Double-encoding (`%2e%2e`) | Reject `%` |

### 5. Skill Files
- Ship skill files or agent documentation with agent-specific invariants
- Encode rules not obvious from `--help`

### 6. Multi-Surface Support
- Same binary serves CLI, MCP (stdio JSON-RPC), extensions, env var auth

### 7. Safety Mechanisms: dry-run + Sanitization
- `--dry-run`: validate locally without hitting APIs
- Response sanitization: defend against prompt injection in API responses

## Recommended Retrofit Order

1. `--output json` (minimum viable machine-readable output)
2. Input validation (assume adversarial input)
3. Schema / `--describe` command
4. Field masks / `--fields`
5. `--dry-run`
6. Skill files / agent documentation
7. MCP support

## Review Checklist

### Structured Output
- [ ] `--output json` flag exists
- [ ] JSON is default in non-TTY
- [ ] Empty lists return `[]` (not a friendly text message)

### Introspection
- [ ] `describe` command or schema output exists

### Input Validation
- [ ] Path traversal rejected
- [ ] Control characters rejected
- [ ] `?#%` rejected
- [ ] Validation happens at external input boundaries

### Skill Files
- [ ] Agent workflow is documented
- [ ] Invariants not obvious from `--help` are stated explicitly

### Safety
- [ ] `--dry-run` exists on mutating commands
- [ ] Interactive UI does not launch in non-TTY environments
