# M0 Windows Terminal interactive acceptance — execution sheet

**Status: not run.** The [unsigned development preview](WINDOWS_PREVIEW.md) passed Windows Server 2025 CI. CI cannot establish interactive Windows 10/11 input-method and focus behavior. This sheet is an execution protocol, not proof of acceptance.

## Preconditions

1. Use a **Windows 10 or 11 desktop**, open Windows Terminal with PowerShell 7. Note the Windows Terminal build, font, scaling, keyboard layout and Chinese IME. Do not disable endpoint security.
2. Download the exact preview artifact and its `manifest.json` from [the successful run](https://github.com/liuhui-code/agent-session-desk/actions/runs/44882270556/artifacts/14698263784). Obtain the official psmux v3.4.0 binary and verify its SHA-256. Use a dedicated temporary test directory; never attach to or stop existing user sessions.
3. In a checkout of the same repository, run this **read-only preflight** with absolute paths. It checks product/psmux SHA-256 and records only non-sensitive environment metadata; it does not create sessions or claim desktop tests passed:

```powershell
./tools/ci/windows-interactive-preflight.ps1 `
  -ProductBinary 'C:\test\asd-windows-amd64.exe' `
  -PsmuxBinary 'C:\test\psmux.exe' `
  -PreviewManifest 'C:\test\manifest.json'
```

4. To test session input, build `cmd/asd-fixture` from the **same source commit** (Go 1.27.4), or use a separate non-sensitive interactive test CLI. The development preview archive does not include the fixture. Never substitute a production agent with credentials when capturing a screen recording.

## Execution steps and observable pass criteria

Record each row as **pass / fail / blocked / not_run**, with timestamp, actual output and a sanitized screenshot or recording. A test is not passed merely because the command returned success.

| ID | Action in Windows Terminal | Required observation |
| --- | --- | --- |
| WIN-UI-01 | Launch manager at 120×32 then narrow to 60×18 and expand | No panic, keycap overflow, broken cursor, or inaccessible selection; preview degrades cleanly |
| WIN-UI-02 | Create two isolated fixture sessions using `asd start`; verify `asd list --json` | Distinct Session/Run identities; correct names and UTF-8 |
| WIN-UI-03 | Select first session, Enter, type ASCII and Chinese text through IME composition/commit | Text reaches only selected fixture, no unintended manager shortcuts |
| WIN-UI-04 | In attached session, compose then cancel Chinese IME, paste multiline Unicode | Cancelled composition not submitted; committed text and paste boundaries match fixture record |
| WIN-UI-05 | While attached, use Ctrl+b then d to detach | Returns to same manager, cursor and keyboard focus usable, session identity unchanged |
| WIN-UI-06 | Switch to second session, enter input, detach, return to first | No cross-session input, stale preview, or wrong Run |
| WIN-UI-07 | Resize while attached, then detach; test Ctrl+C with disposable fixture | Resizing/cursor correct; signal affects intended target, manager remains operable |
| WIN-UI-08 | Quit manager with q, reopen using identical backend/runtime/namespace | Both background Run identities unchanged; no implicit restart |
| WIN-UI-09 | Explicitly stop both sessions using their exact Session IDs | Correct processes terminate; no unrelated user session affected |
| WIN-UI-10 | Repeat with Windows Terminal alternate font, display scaling and IME | Differences recorded; do not generalize unsupported combinations |

Suggested isolated commands (replace paths and use the same options on **every** invocation):

```powershell
$asd = 'C:\test\asd-windows-amd64.exe'
$backend = 'C:\test\psmux.exe'
$fixture = 'C:\test\asd-fixture.exe'
$runtime = Join-Path $env:TEMP 'asd-manual-test-20261121'
$common = @('--backend', $backend, '--runtime-dir', $runtime, '--namespace', 'asd-manual')
& $asd @common doctor
& $asd @common start --name '中文输入 A' --cwd $PWD.Path -- $fixture
& $asd @common start --name '会话 B' --cwd $PWD.Path -- $fixture
& $asd @common list --json
& $asd @common
```

Use the exact `session_id` values returned by `list --json` for `preview`, `attach`, and `stop`. Do not guess or copy a different Run's ID.

## Reporting

Create an execution record with: preview source commit, hashes, Windows/Terminal/IME versions, per-case result, steps to reproduce, sanitized screenshots, known limitations, and links to related Feature Ledger entries FL-002, FL-003, FL-006, FL-012, FL-016. Do **not** upload `PSMUX_DATA_DIR`, environment dumps, agent transcripts, authentication files, or user home directories. `preflight.json` deliberately excludes them.

**Acceptance gate:** all mandatory cases above must have actual evidence on a supported Windows desktop. Failed/blocked/not-run cases keep M0 and FL-016 open. CI success alone never closes this gate.
