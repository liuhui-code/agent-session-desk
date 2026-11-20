# Repository working notes

## Current state

Updated 2026-11-21. Architecture and technology directions have been accepted and recorded; the product is not implemented, no executable release exists, and native Windows behavior has not been tested. The project license, exact reuse scope, SQLite driver, IPC transport, and measured resource budgets still need resolution.

The implementation target is Go 1.27 with stable Charm v2 libraries (Bubble Tea 2.1.2, Lip Gloss 2.1.1, Bubbles 2.2.2 only where needed). Reuse selected Agent Deck UI code and Agent Manager's Windows adaptation code through explicit local interfaces. Upstream v1 source is a migration reference, not the intended release dependency baseline. Windows uses the psmux 3.4.0 stable release as an M0 validation candidate; do not equate newer master features with that release. The later Unix backend is tmux.

Read README.md, docs/TECH_STACK.md, docs/PRODUCT_PLAN.md, docs/ROADMAP.md, docs/DECISIONS.md, docs/adr/README.md, and docs/FEATURE_LEDGER.md before implementation work. Exact dependency and upgrade policy is owned by ADR-0005 and TECH_STACK.md; acceptance of a direction does not establish compatibility.

## Product invariants

- Target native Windows; WSL is not a core prerequisite.
- Launch local custom agents through CLI commands. ACP, MCP, HTTP, and custom SDKs are optional enhancements, not entry requirements.
- Preserve a minimal multi-session TUI using Agent Deck's suitable layout and interaction code. Prefer one session list, the selected preview, and on-demand dialogs. The v0.1 design images are unaccepted exploration, not the implementation baseline.
- Start with list/preview management and full-screen attach. The manager sidebar and tabs are hidden during attach; return to the manager after detaching. A permanently embedded interactive terminal is deferred.
- Keep Session, Run, and Tab identities distinct.
- Separate terminal output from structured conversation history and model context.
- Separate queued, input-written, agent-accepted, and completed message states. Never infer stronger states without evidence.
- Unknown CLI readiness must remain unknown. Automatic queued delivery requires a reliable readiness signal.
- Reattaching to a live process is different from restoring an exited agent's conversation.
- A live multiplexer session or shell does not prove the agent itself is alive. Keep terminal, Run, agent-exit evidence, and agent work state distinct.
- In M0-M2, psmux owns the terminal process lifetime; the shared product core persists metadata and messages, and a product-owned resident Host is not mandatory. M3 background automatic queue consumption uses one shared lightweight Host when that capability is enabled. Queue persistence alone does not imply automatic delivery.
- Keep resources bounded: TOML for configuration, SQLite for metadata/messages, and bounded files for output records. Disable unnecessary warm processes in the product's own runtime configuration, avoid duplicate full histories, and refresh previews on demand.

## Working conventions

- Use Chinese for product documentation unless the task asks otherwise.
- Treat recommendations and open decisions as proposals until evidence or an explicit product decision resolves them.
- Follow accepted ADRs while preserving their unresolved validation gates. Record a superseding ADR when changing an accepted architectural decision.
- Keep requirement IDs and relative documentation links stable.
- Record upstream sources, exact versions/commits when available, and license findings before code reuse.
- Adapt migrated v1 UI code to the stable v2 dependency baseline, then verify the affected input, layout, terminal restoration, and Windows behavior. Do not silently turn an older source dependency into a release dependency just because the upstream snapshot used it.
- Do not claim native Windows compatibility based only on Linux execution, a mockup, or a framework support statement.
- Keep credentials, private session content, and local runtime data out of commits.
- Prefer concrete, reproducible acceptance scenarios over speculative architecture expansion.
- Keep Feature Ledger IDs and FR mappings stable. Update actual implementation, evidence, and release status with each relevant change; documents and mockups do not count as implemented or Windows-validated features.
- Summarize what changed, why, how it was verified, and remaining limits.
