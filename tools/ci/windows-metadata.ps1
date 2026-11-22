[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$ProductBinary,
    [Parameter(Mandatory=$true)][string]$FixtureBinary,
    [Parameter(Mandatory=$true)][string]$PsmuxBinary,
    [Parameter(Mandatory=$true)][string]$EvidenceDirectory,
    [string]$StoreTestBinary = '',
    [string]$ConfigScript = ''
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$PSNativeCommandUseErrorActionPreference=$false
if (-not $IsWindows) { throw 'Native Windows only.' }
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$product=(Resolve-Path -LiteralPath $ProductBinary).Path
$root=[IO.Path]::GetFullPath($EvidenceDirectory)
if (Test-Path -LiteralPath $root) {throw 'Use a fresh evidence directory.'}
New-Item -ItemType Directory -Path $root | Out-Null
function Write-Utf8([string]$Path,[string]$Text){[IO.File]::WriteAllText($Path,$Text,[Text.UTF8Encoding]::new($false))}
$failure=$null
$checks=[Collections.Generic.List[object]]::new()
try {
    $fixtureCase=Join-Path $root 'fixture-case'
    if (-not $ConfigScript){$ConfigScript=Join-Path $PSScriptRoot 'windows-config.ps1'}
    $configOutput=& ([ScriptBlock]::Create([IO.File]::ReadAllText($ConfigScript))) -ProductBinary $product -FixtureBinary $FixtureBinary -PsmuxBinary $PsmuxBinary -EvidenceDirectory $fixtureCase
    Write-Utf8 (Join-Path $root 'fixture-case.log') ($configOutput | Out-String)
    $runtime=Join-Path $fixtureCase 'runtime'
    $namespace='asd-config-native'
    $history=& $product --backend (Join-Path $root 'absent-backend.exe') --runtime-dir $runtime --namespace $namespace history 2>(Join-Path $root 'history.stderr')
    if ($LASTEXITCODE -ne 0){throw 'Offline history failed.'}
    Write-Utf8 (Join-Path $root 'history.json') ($history | Out-String)
    $rows=@($history | ConvertFrom-Json)
    $observed=[IO.File]::ReadAllText((Join-Path $fixtureCase 'result.json')) | ConvertFrom-Json
    if ($rows.Count -ne 2){throw 'Wrong history count.'}
    foreach ($row in $rows){
        $match=@($observed.observations | Where-Object { $_.scenario -in @('template','override') -and $_.session_id -eq $row.session.session_id -and $_.run_id -eq $row.session.run_id })
        if ($match.Count -ne 1 -or $row.session.terminal_state -ne 'unknown' -or -not $row.stop_confirmed_at){throw 'Historical identity/liveness/stop evidence changed.'}
    }
    $checks.Add(@{name='reopen/offline-history/confirmed-stop';result='passed';count=$rows.Count})
    $state=Join-Path $runtime $namespace
    $other=& $product --runtime-dir $runtime --state-dir $state --namespace asd-other history
    if ($LASTEXITCODE -ne 0 -or @($other|ConvertFrom-Json).Count -ne 0){throw 'Namespace data leaked.'}
    $checks.Add(@{name='namespace-isolation';result='passed'})
    $one=& $product --runtime-dir $runtime --namespace $namespace history --limit 1
    if ($LASTEXITCODE -ne 0 -or @($one|ConvertFrom-Json).Count -ne 1){throw 'History limit ignored.'}
    $bad=& $product --runtime-dir $runtime --namespace $namespace history --limit 0 2>(Join-Path $root 'limit.stderr')
    if ($LASTEXITCODE -eq 0){throw 'Invalid history limit accepted.'}
    $checks.Add(@{name='history-limit';result='passed'})
    $list=& $product --backend $PsmuxBinary --runtime-dir $runtime --namespace $namespace list --json
    if ($LASTEXITCODE -ne 0 -or @($list|ConvertFrom-Json).Count -ne 0){throw 'Historical metadata became live sessions.'}
    $checks.Add(@{name='live-list-empty';result='passed'})
    if ($StoreTestBinary){
        $test=(Resolve-Path $StoreTestBinary).Path
        $output=& $test '-test.v' 2>&1 | Out-String
        $code=$LASTEXITCODE
        Write-Utf8 (Join-Path $root 'store-tests.log') $output
        if ($code -ne 0 -or $output.Contains('SKIP')){throw 'Native SQLite tests failed or skipped.'}
        $checks.Add(@{name='native-SQLite-tests';result='passed';sha256=(Get-FileHash $test -Algorithm SHA256).Hash.ToLowerInvariant()})
    }
} catch {$failure=$_.Exception.Message} finally {
    Write-Utf8 (Join-Path $root 'result.json') (ConvertTo-Json -Depth 8 -InputObject ([ordered]@{
        utc=[DateTime]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString
        product_sha256=(Get-FileHash $product -Algorithm SHA256).Hash.ToLowerInvariant()
        checks=$checks.ToArray();failure=$failure
    }))
}
if ($failure){throw $failure}
Write-Output 'Native metadata persistence and history: passed.'
