# Repository working notes

## Current state

This is the planning repository for Agent Session Desk. No executable product, implementation language, UI base, terminal backend, or project license has been selected yet. Read README.md, docs/PRODUCT_PLAN.md, docs/ROADMAP.md, and docs/DECISIONS.md before implementation work.

## Product invariants

- Target native Windows; WSL is not a core prerequisite.
- Launch local custom agents through CLI commands. ACP, MCP, HTTP, and custom SDKs are optional enhancements, not entry requirements.
- Preserve a polished multi-session TUI experience inspired by Agent Deck / Agent of Empires.
- Keep Session, Run, and Tab identities distinct.
- Separate terminal output from structured conversation history and model context.
- Separate queued, input-written, agent-accepted, and completed message states. Never infer stronger states without evidence.
- Unknown CLI readiness must remain unknown. Automatic queued delivery requires a reliable readiness signal.
- Reattaching to a live process is different from restoring an exited agent's conversation.

## Working conventions

- Use Chinese for product documentation unless the task asks otherwise.
- Treat recommendations and open decisions as proposals until evidence or an explicit product decision resolves them.
- Keep requirement IDs and relative documentation links stable.
- Record upstream sources, exact versions/commits when available, and license findings before code reuse.
- Do not claim native Windows compatibility based only on Linux execution, a mockup, or a framework support statement.
- Keep credentials, private session content, and local runtime data out of commits.
- Prefer concrete, reproducible acceptance scenarios over speculative architecture expansion.
- Summarize what changed, why, how it was verified, and remaining limits.
