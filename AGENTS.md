# Repository working notes

## Current state

Updated 2026-11-21. The first M0 source migration slice is implemented: Go/Charm v2 management UI, a narrow shared Session/Run model, custom CLI commands, and isolated experimental psmux/tmux backends. At commit b6f2cd2adf24a77d52244aeed1afd1c2ab5dad87, GitHub run 44860102659 passed both CI jobs. On its Windows Server 2025 runner, all five real-backend tests and six rejection subtests passed with no native failures or skips; ordinary Go tests, race, vet, catalog checks and native builds also passed. Exact environment, source hashes and earlier failed attempts are preserved in docs/testing/results/2026-11-21-m0-native-windows.md, alongside the original Linux evidence in docs/testing/results/2026-11-21-m0-source-migration.md. The next short native resource slice passed at commit 30eaf058ed507fe8fea623593273ba6abf0162dc, GitHub run 44865161425: seven scenarios, 22 phases, 280 snapshots and seven normal cleanups. It measures the actual TUI at 0/1/5/10 sessions, controlled finite output, UI close/reopen and direct-pipe controls. Final counts and all four raw artifacts are archived in docs/testing/results/2026-11-21-m0-windows-resources.md and its evidence directory. These are bounded process/backend contracts, not an accepted release, complete M0 acceptance or Windows desktop compatibility claim. Read docs/M0_DEVELOPMENT.md and docs/REUSE_MANIFEST.md for implementation scope. The original-project license, remaining reuse scope, SQLite driver, IPC transport and measured resource budgets still need resolution.

The implementation target is Go 1.27 with stable Charm v2 libraries (Bubble Tea 2.1.2, Lip Gloss 2.1.1, Bubbles 2.2.2 only where needed). Reuse selected Agent Deck UI code and Agent Manager's Windows adaptation code through explicit local interfaces. Upstream v1 source is a migration reference, not the intended release dependency baseline. Windows uses the psmux 3.4.0 stable release as an M0 validation candidate; do not equate newer master features with that release. The later Unix backend is tmux.

Read README.md, docs/TECH_STACK.md, docs/PRODUCT_PLAN.md, docs/ROADMAP.md, docs/DECISIONS.md, docs/adr/README.md, and docs/FEATURE_LEDGER.md before implementation work. Exact dependency and upgrade policy is owned by ADR-0005 and TECH_STACK.md; acceptance of a direction does not establish compatibility.

Native Windows automation is described in docs/testing/NATIVE_WINDOWS.md. The pinned psmux dependency is checked by archive and executable hashes; the hosted runner is Windows Server 2025 and contains WSL. Opt-in TestWindowsPsmuxNative tests execute real processes and must never be replaced by fake helpers or cross-compilation evidence. Short resource automation is described in docs/testing/WINDOWS_RESOURCES.md. Extended DOS image-path equivalence is deliberately narrow; preserve the exact PID/FILETIME/path, liveness and ownership checks. Live fixture/registry files require shared reads. Observed sums are not complete lifetime costs: the final managed-10/all-output phase has 0/12 complete-observation coverage, and managed-5/recovery-idle has 11/12. Do not remove gaps or fill missing values with zero. Interactive Windows Terminal/IME/attach, WSL-absence, full-duration resource budgets, long soak and release acceptance remain separate gates. A workflow file's presence does not establish that it ran; record the exact run and outcome.

Test design now lives in docs/testing/. Read TEST_STRATEGY.md, COVERAGE.md, and the applicable JSON cases before implementation. All 339 catalog cases retain their designed/not_run baseline; actual partial evidence is stored separately. Go unit/subprocess-contract tests, an independently observed cmd/asd-fixture and the native Windows runner now exist; the complete planned fixtures and acceptance scenarios remain unfinished. The Python catalog tool validates documentation structure and renders derived files only; its success is not product or Windows test evidence. Run go test ./..., go test -race ./..., go vet ./... and go mod verify for the implementation, then the catalog check. A Windows cross-build is not a Windows test execution. M0 stores only bounded live-discovery tags inside the backend, not a replacement metadata database; Explicit read-only TOML launch templates and minimal SQLite Session/Run metadata are implemented; see ADR-0007/0009 and the 2026-11-22 reports. Message storage, migration, output logs and full acceptance remain pending. Known psmux/tmux parser mutations are rejected before launch; do not remove these guards without proving literal transport. Version probing parses the reported release and its recognized build-provenance suffix; only the separate archive/executable hash checks establish which pinned psmux artifact the native runner exercised.

## Product invariants

- Target native Windows; WSL is not a core prerequisite.
- Launch local custom agents through CLI commands. ACP, MCP, HTTP, and custom SDKs are optional enhancements, not entry requirements.
- Preserve a minimal multi-session TUI using Agent Deck's suitable layout and interaction code. Prefer one session list, the selected preview, and on-demand dialogs. The v0.1 design images are unaccepted exploration, not the implementation baseline.
- Start with list/preview management and full-screen attach. The manager sidebar and tabs are hidden during attach; return to the manager after detaching. A permanently embedded interactive terminal is deferred.
- Keep Session, Run, and Tab identities distinct.
- Separate terminal output from structured conversation history and model context.
- Separate queued, input-written, agent-accepted, and completed message states. Never infer stronger states without evidence.
- Unknown CLI readiness must remain unknown. Automatic queued delivery requires a reliable readiness signal.
- Follow ADR-0006 for human attention: verified approval/question/inspection events have separate item identity, lifecycle, read/snooze state, observation freshness, and notification outcome. Viewing or dismissing a notification does not resolve an interaction request; notification activation only navigates and revalidates, never approves or injects input.
- Exiting a Run invalidates its old interactive prompts, but an abnormal-exit inspection item may remain linked to that ended Run. Disconnect alone is uncertainty, not resolution. Automatic queued tasks must not enter an approval prompt.
- Continue event handling independently of TUI rendering during full-screen attach. After UI exit, continued attention monitoring requires a running shared Host or a separately verified hook path; terminal survival alone does not prove it. Desktop, sound, and completion notifications default off; no per-session notifier daemon or unbounded event history.
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
- Keep TC IDs stable and edit test content in docs/testing/cases/*.json. Regenerate derived Markdown/CSV with tools/test_catalog.py, then check --generated. Record unresolved behaviors and budgets instead of silently inventing acceptance rules; save actual run evidence separately from design coverage.
- Summarize what changed, why, how it was verified, and remaining limits.

## Development guardrails

- 使用 TDD：新增行为或修复按一条行为 RED → GREEN → 必要重构推进，保存关键失败和复验依据；重构沿用行为测试。
- 每个相关变更同步 ADR、Feature Ledger、实际测试报告和 Git 提交记录。架构决定的接受不等于实现或验收完成。
- 产品代码正文单文件不得超过 500 行；生产 Go、Python 与 PowerShell 文件采用更严格的物理总行数检查（含注释和空行；Go 测试文件除外），由 `python3 -B tools/ci/check_source_limits.py` 执行。测试文件单独组织，新增文件也保持短小。
- UI 样式 100% 由 design tokens 管理：颜色、粗细、间距、装饰符与响应式尺寸统一定义在 `internal/ui/design_tokens.go`；`styles.go` 仅组合 token，视图不得新增独立样式数值。单元格对齐、数组下标与固定结构行数不是设计值。新增 UI 必须沿用该入口并验证渲染边界。
