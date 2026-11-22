[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string]$ProductBinary,
    [Parameter(Mandatory = $true)][string]$FixtureBinary,
    [Parameter(Mandatory = $true)][string]$PsmuxBinary,
    [Parameter(Mandatory = $true)][string]$EvidenceDirectory
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
if ([Environment]::OSVersion.Platform -ne 'Win32NT') { throw 'Native Windows only.' }
[Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
$product = (Resolve-Path -LiteralPath $ProductBinary).Path
$fixture = (Resolve-Path -LiteralPath $FixtureBinary).Path
$psmux = (Resolve-Path -LiteralPath $PsmuxBinary).Path
if ((Get-FileHash $psmux -Algorithm SHA256).Hash.ToLowerInvariant() -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca') { throw 'Wrong pinned psmux artifact.' }
$root = [IO.Path]::GetFullPath($EvidenceDirectory)
if (Test-Path -LiteralPath $root) { throw 'Use a fresh evidence directory.' }
New-Item -ItemType Directory -Path $root | Out-Null
$project = Join-Path $root 'project 中文'
New-Item -ItemType Directory -Path $project | Out-Null
$config = Join-Path $root 'agents.toml'
$globalArgs = @('--backend', $psmux, '--runtime-dir', (Join-Path $root 'runtime'), '--namespace', 'asd-config-native')
$commands = [Collections.Generic.List[object]]::new()
$observations = [Collections.Generic.List[object]]::new()
$sessions = [Collections.Generic.List[string]]::new()
$identities = [Collections.Generic.List[object]]::new()
$failure = $null
function Write-Utf8([string]$Path, [string]$Text) { [IO.File]::WriteAllText($Path, $Text, [Text.UTF8Encoding]::new($false)) }
function Invoke-Product([string[]]$Arguments, [switch]$Reject) {
    $errPath = Join-Path $root ('stderr-' + $commands.Count + '.txt')
    $text = (& $product @globalArgs @Arguments 2>$errPath | Out-String)
    $code = $LASTEXITCODE
    $stderr = [IO.File]::ReadAllText($errPath)
    $commands.Add([ordered]@{args=$Arguments; code=$code; stdout=$text; stderr=$stderr})
    if (($Reject -and $code -eq 0) -or (-not $Reject -and $code -ne 0)) { throw "Unexpected CLI exit $code : $stderr" }
    return $text
}
function Read-Shared([string]$Path) {
    $stream = [IO.File]::Open($Path, 'Open', 'Read', [IO.FileShare]::ReadWrite)
    $reader = [IO.StreamReader]::new($stream, [Text.Encoding]::UTF8)
    try { return $reader.ReadToEnd() } finally { $reader.Dispose() }
}
function Start-Template([string]$Name, [string[]]$Overrides) {
    $report = Join-Path $root ($Name + '.jsonl')
    $body = @"
schema_version = 1
[templates.dev]
name = '模板'
executable = '$fixture'
args = ['', 'a b', '中文', 'a"b', 'end\']
cwd = 'project 中文'
[templates.dev.env]
ASD_CONFIG_VALUE = 'template value'
ASD_FIXTURE_ENV_KEYS = 'ASD_CONFIG_VALUE'
ASD_FIXTURE_REPORT_PATH = '$report'
"@
    Write-Utf8 $config $body
    $result = Invoke-Product (@('--config', $config, 'start', '--template', 'dev') + $Overrides) | ConvertFrom-Json
    $sessions.Add($result.session_id)
    $deadline = [DateTime]::UtcNow.AddSeconds(15)
    $started = $null
    while ([DateTime]::UtcNow -lt $deadline) {
        if (Test-Path $report) {
            $rows = @((Read-Shared $report) -split "`n" | Where-Object { $_.EndsWith('}') })
            if ($rows.Count -gt 0) { $started = $rows[0] | ConvertFrom-Json; break }
        }
        Start-Sleep -Milliseconds 100
    }
    if ($null -eq $started -or $started.event -ne 'started') { throw 'Fixture start not observed.' }
    $process = Get-Process -Id $started.pid -ErrorAction Stop
    $identities.Add([ordered]@{pid=$started.pid; start_filetime=$process.StartTime.ToUniversalTime().ToFileTimeUtc(); path=$process.Path})
    if ($process.Path -ne $fixture) { throw 'Fixture process ownership mismatch.' }
    $wantArgs = @('', 'a b', '中文', 'a"b', 'end\')
    if ($started.args.Count -ne $wantArgs.Count) { throw 'Wrong argv length.' }
    for ($i=0; $i -lt $wantArgs.Count; $i++) { if ($started.args[$i] -cne $wantArgs[$i]) { throw "argv $i changed." } }
    $wantDir = $project; $wantValue = 'template value'; $wantTitle = '模板'
    if ($Overrides.Count -gt 0) { $wantDir = $root; $wantValue = 'CLI value'; $wantTitle = '覆盖' }
    if ($started.cwd -ne $wantDir -or $started.env.ASD_CONFIG_VALUE -cne $wantValue -or $result.title -cne $wantTitle) { throw 'cwd/env/title changed.' }
    $observations.Add([ordered]@{scenario=$Name; session_id=$result.session_id; run_id=$result.run_id; title=$result.title; fixture=$started; result='passed'})
}
try {
    Start-Template 'template' @()
    Start-Template 'override' @('--cwd', $root, '--name', '覆盖', '--env', 'ASD_CONFIG_VALUE=CLI value')
    $before = @(Invoke-Product @('list','--json') | ConvertFrom-Json)
    if ($before.Count -ne 2 -or $before[0].session_id -eq $before[1].session_id -or $before[0].run_id -eq $before[1].run_id) { throw 'Session/Run identity collision.' }
    foreach ($body in @('schema_version=2', "schema_version=1`n[templates.dev]`nexecutable='$fixture'`nargz='synthetic-sensitive-marker'")) {
        Write-Utf8 $config $body
        $null = Invoke-Product @('--config',$config,'start','--template','dev') -Reject
        if ($commands[$commands.Count-1].stderr.Contains('synthetic-sensitive-marker')) { throw 'Configuration value leaked.' }
    }
    $null = Invoke-Product @('--config',$config,'start','--template','dev','--',$fixture) -Reject
    $after = @(Invoke-Product @('list','--json') | ConvertFrom-Json)
    if ($after.Count -ne 2) { throw 'Rejected configuration created a session.' }
    $observations.Add([ordered]@{scenario='rejections'; count=3; result='passed'})
} catch { $failure = $_.Exception.Message } finally {
    foreach ($id in $sessions) {
        try { $null = Invoke-Product @('stop', $id) } catch { $failure = "Cleanup failed: $($_.Exception.Message); $failure" }
    }
    foreach ($identity in $identities) {
        $process = Get-Process -Id $identity.pid -ErrorAction SilentlyContinue
        if ($null -ne $process -and $process.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $identity.start_filetime) { $failure = "Owned fixture still alive: $($identity.pid); $failure" }
    }
    Write-Utf8 (Join-Path $root 'commands.json') (ConvertTo-Json -InputObject @($commands.ToArray()) -Depth 8)
    Write-Utf8 (Join-Path $root 'result.json') (ConvertTo-Json -Depth 10 -InputObject ([ordered]@{
        utc=[DateTime]::UtcNow.ToString('o'); os=[Environment]::OSVersion.VersionString
        product_sha256=(Get-FileHash $product -Algorithm SHA256).Hash.ToLowerInvariant()
        fixture_sha256=(Get-FileHash $fixture -Algorithm SHA256).Hash.ToLowerInvariant()
        psmux_sha256=(Get-FileHash $psmux -Algorithm SHA256).Hash.ToLowerInvariant()
        observations=$observations.ToArray(); identities=$identities.ToArray(); failure=$failure
    }))
}
if ($failure) { throw $failure }
Write-Output 'Native TOML configuration: passed; owned sessions and fixture processes cleaned.'
