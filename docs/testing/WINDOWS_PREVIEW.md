# Windows x64 development preview

This is an **unsigned, manually triggered development artifact**, not a release. The repository's original-code license has not yet been decided, so no GitHub Release, package manager publication, or automatic binary distribution is configured.

## Build and verification

[Windows development preview workflow](../../.github/workflows/windows-preview.yml) runs on native Windows Server 2025 using the pinned Go toolchain and GitHub Actions revisions. It builds `cmd/asd` with `CGO_ENABLED=0`, `-trimpath` and `-buildvcs=true`, then runs the existing native Windows verification against the SHA-256-pinned psmux 3.4.0 binary. **Only a successful build and native verification upload the preview artifact.** Failed runs still upload verification evidence when present.

To request a build, open GitHub Actions → **Windows development preview** → **Run workflow** on the intended commit/branch. Download the resulting `asd-windows-amd64-dev-<sha>` artifact. It contains:

- `asd-windows-amd64.exe`: unsigned product binary
- `SHA256SUMS`: checksum of that exact executable
- `manifest.json`: source SHA, target platform, toolchain, dependency requirement, checksum, and scope

The workflow intentionally does not bundle or redistribute psmux. Obtain the **official psmux v3.4.0** Windows x64 release using [the verified dependency installer](../../tools/ci/install-psmux.ps1), or independently verify the official executable SHA-256 against [windows-dependencies.json](../../tools/ci/windows-dependencies.json). Keep `psmux.exe` available at the configured backend path. Do not download similarly named unverified binaries.

## Manual acceptance still required

Run in Windows 10/11 **Windows Terminal**, not just the Server 2025 CI environment. Use the [M0 development guide](../M0_DEVELOPMENT.md) to create a controlled CLI session. Record:

1. Windows build, Windows Terminal version, keyboard layout, and psmux SHA-256.
2. Chinese IME composition/commit/cancel while a session is attached; no unintended manager shortcut.
3. Attach and detach, then verify the same Session/Run remains and the manager regains focus.
4. Terminal resize at narrow and wide dimensions; verify no truncation, misplaced cursor, or keycap overflow.
5. Quit and reopen the manager, confirm the background session remains; stop it explicitly and verify cleanup.

These are **not** automatically passed by the workflow. Preserve failed steps, screen recordings or sanitized logs, and record actual evidence in Feature Ledger before changing acceptance status.

The artifact is ephemeral (14 days), unsigned and may trigger Windows security prompts. Do not bypass system protection or recommend disabling antivirus. No installation or auto-update is implied.
