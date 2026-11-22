$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$path=Join-Path $base 'windows-history-resize.ps1'
$tokens=$null;$errors=$null;$text=[IO.File]::ReadAllText($path)
$null=[Management.Automation.Language.Parser]::ParseInput($text,[ref]$tokens,[ref]$errors)
if($errors.Count){$errors|ForEach-Object{$_.Message};throw 'Parser errors'}
& ([ScriptBlock]::Create($text)) -ProductBinary (Join-Path $base 'copy-final2-product.exe') -ProductSHA256 a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0 -PsmuxBinary (Join-Path $base 'psmux-history-candidate.exe') -PsmuxSHA256 ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937 -ShellBinary (Join-Path $base 'pwsh7/pwsh.exe') -ArtifactRoot (Join-Path $base 'history-tool-candidate1')
