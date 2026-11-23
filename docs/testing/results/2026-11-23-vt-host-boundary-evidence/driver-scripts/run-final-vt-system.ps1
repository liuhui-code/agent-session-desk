$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$source=[IO.File]::ReadAllText((Join-Path $base 'windows-vt-boundary.ps1'))
& ([ScriptBlock]::Create($source)) -PsmuxBinary (Join-Path $base 'history-backend-candidate/psmux.exe') -PsmuxSHA256 ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937 -ShellBinary (Join-Path $base 'pwsh7/pwsh.exe') -ArtifactRoot (Join-Path $base 'vt-final-system') -BeforeAttachSeed
