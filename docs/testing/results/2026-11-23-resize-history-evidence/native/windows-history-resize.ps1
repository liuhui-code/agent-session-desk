param(
    [Parameter(Mandatory)][string]$ProductBinary,
    [Parameter(Mandatory)][string]$ProductSHA256,
    [Parameter(Mandatory)][string]$PsmuxBinary,
    [string]$PsmuxSHA256 = 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca',
    [Parameter(Mandatory)][string]$ShellBinary,
    [Parameter(Mandatory)][string]$ArtifactRoot,
    [ValidateSet('BeforeAttach', 'AfterAttach')][string]$OutputTiming = 'BeforeAttach',
    [switch]$DirectBackend,
    [switch]$IdleWriter,
    [switch]$SkipSecondaryResize,
    [switch]$SkipEarlyCopy,
    [switch]$ObserveOnly
)
# This opt-in native diagnostic creates one synthetic session in a fresh root.
# Default exit status asserts complete retained history, not just a visible page.
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
foreach ($path in @($ProductBinary, $PsmuxBinary, $ShellBinary, $ArtifactRoot)) {
    if (![IO.Path]::IsPathRooted($path)) { throw 'All paths must be absolute' }
}
if (Test-Path $ArtifactRoot) { throw 'ArtifactRoot must not exist' }
$hashes = @{}
foreach ($pair in @(@($ProductBinary, $ProductSHA256), @($PsmuxBinary, $PsmuxSHA256), @($ShellBinary, ''))) {
    $digest = (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($pair[1] -and $digest -ne $pair[1].ToLowerInvariant()) { throw 'Binary SHA-256 mismatch' }
    $hashes[[IO.Path]::GetFileName($pair[0])] = [string]$digest
}
$root = [IO.Path]::GetFullPath($ArtifactRoot)
$null = New-Item -ItemType Directory -Path $root
$ns = 'asd-history-' + [Guid]::NewGuid().ToString('N').Substring(0, 12)
$runtime = Join-Path $root 'runtime'; $data = Join-Path $runtime $ns
$g = @('--backend', $PsmuxBinary, '--runtime-dir', $runtime, '--namespace', $ns, '--state-dir', (Join-Path $root 'state'))
$fixture = Join-Path $root 'history-fixture.ps1'
@'
function global:prompt { '> ' }
$p = Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content identity.json -Encoding utf8
'BOOT_READY'
while (!(Test-Path gate.txt)) { Start-Sleep -Milliseconds 20 }
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40 | ForEach-Object { 'history line ' + $_ }
'HISTORY_END'
'done' | Set-Content done.txt
if (Test-Path idle-writer.txt) { while ($true) { Start-Sleep -Milliseconds 100 } }
'@ | Set-Content -LiteralPath $fixture -Encoding utf8
$s = $null; $tcp = $null; $identity = $null; $pane = $null; $failure = $null
$stages = [Collections.Generic.List[object]]::new()
$captures = [Collections.Generic.List[object]]::new()
$result = @{result='started';hashes=$hashes;namespace=$ns;utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;output_timing=$OutputTiming;direct_backend=[bool]$DirectBackend;idle_writer=[bool]$IdleWriter;secondary_resize=(!$SkipSecondaryResize);early_copy=(!$SkipEarlyCopy)}
function Send([string]$line) {
    $bytes = [Text.Encoding]::UTF8.GetBytes($line + "`n")
    $stream.Write($bytes, 0, $bytes.Length); $stream.Flush()
}
function Hex([string]$text) {
    $hex = ([Text.Encoding]::UTF8.GetBytes($text) | ForEach-Object { $_.ToString('x2') }) -join ' '
    Send "send-keys -t %$pane -H $hex"
}
function Lines($f) { return @($f.layout.rows_v2 | ForEach-Object { ($_.runs.text -join '').TrimEnd() }) }
function WaitFrame([scriptblock]$predicate) {
    $until = [DateTime]::UtcNow.AddSeconds(12)
    while ([DateTime]::UtcNow -lt $until) {
        Send 'dump-state'; $line = $reader.ReadLine()
        if ($null -eq $line) { throw 'Terminal stream EOF' }
        if (!$line.StartsWith('{')) {
            if ($line -eq 'NC' -or $line.Trim() -eq '') { continue }
            throw 'Unexpected terminal response'
        }
        if ([Text.Encoding]::UTF8.GetByteCount($line) -gt 2MB) { throw 'Frame exceeds diagnostic bound' }
        $f = $line | ConvertFrom-Json
        if ($f.session_name -ne $s.backend_id -or $f.layout.type -ne 'leaf' -or ($null -ne $pane -and $f.layout.id -ne $pane)) { throw 'Session/pane changed' }
        if (& $predicate $f) { return $f }
        Start-Sleep -Milliseconds 30
    }
    throw 'Expected frame not observed within 12 seconds'
}
function SaveStage([string]$label, $f) {
    $p = $f.layout
    $stages.Add(@{label=$label;cols=$p.cols;rows=$p.rows;copy_mode=$p.copy_mode;scroll_offset=$p.scroll_offset;view_offset=$p.view_offset;copy_cursor_row=$p.copy_cursor_row;copy_cursor_col=$p.copy_cursor_col;status_message=$f.status_message;rows_v2_text=@(Lines $f);content_text=@($p.content | ForEach-Object { ($_.text -join '').TrimEnd() });rows_v2=$p.rows_v2;content=$p.content})
}
function Capture([string]$label) {
    if ($DirectBackend) {
        $text = (& $PsmuxBinary -L $ns capture-pane -t ($s.backend_id + ':.%' + $pane) -p -S -60 | Out-String)
        if ($LASTEXITCODE -ne 0) { throw 'Direct capture failed' }
        $snapshot = @{text=$text;truncated='unknown'}
    } else {
        $v = (& $ProductBinary @g preview $s.session_id | Out-String) | ConvertFrom-Json
        if ($LASTEXITCODE -ne 0 -or $v.session_id -ne $s.session_id -or $v.run_id -ne $s.run_id) { throw 'Preview failed or identity changed' }
        $snapshot = $v.snapshot; $text = $snapshot.text
    }
    $ids = @(([string]$text -split '\r?\n') | ForEach-Object { if ($_ -match '^history line (\d+)$') { [int]$Matches[1] } })
    $complete = (($ids -join ',') -ceq ((1..40) -join ','))
    $captures.Add(@{label=$label;snapshot=$snapshot;ids=$ids;complete=$complete})
}
try {
    if ($IdleWriter) { [IO.File]::WriteAllText((Join-Path $root 'idle-writer.txt'), 'idle') }
    if ($DirectBackend) {
        # This script runs in its own PowerShell process; never source it into a
        # user's shell. Native backend children inherit only this private route.
        foreach ($entry in @(Get-ChildItem Env: | Where-Object { $_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX', 'TMUX_PANE', 'TMUX_TMPDIR') })) { Remove-Item ('Env:' + $entry.Name) }
        $env:PSMUX_NO_WARM='1'; $env:PSMUX_ALLOW_NESTING='1'; $env:PSMUX_DATA_DIR=$data
        $null = New-Item -ItemType Directory -Path $data
        $config = Join-Path $root 'psmux.conf'
        [IO.File]::WriteAllText($config, "set -g warm off`nset -g status off`nset -g history-limit 2000`n")
        $env:PSMUX_CONFIG_FILE=$config; $backendID='history-' + [Guid]::NewGuid().ToString('N')
        & $PsmuxBinary -L $ns new-session -d -s $backendID -c $root -x 120 -y 32 -- $ShellBinary -NoLogo -NoProfile -NoExit -File $fixture | Out-Null
        if ($LASTEXITCODE -ne 0) { throw 'Direct start failed' }
        $s=[pscustomobject]@{backend_id=$backendID;session_id='direct-diagnostic';run_id='direct-diagnostic'}
    } else {
        $s = (& $ProductBinary @g start --name 'Synthetic history resize' --cwd $root -- $ShellBinary -NoLogo -NoProfile -NoExit -File $fixture | Out-String) | ConvertFrom-Json
        if ($LASTEXITCODE -ne 0) { throw 'Product start failed' }
    }
    $result.session=$s
    $until=[DateTime]::UtcNow.AddSeconds(10)
    while (!(Test-Path (Join-Path $root 'identity.json')) -and [DateTime]::UtcNow -lt $until) { Start-Sleep -Milliseconds 50 }
    $identity=[IO.File]::ReadAllText((Join-Path $root 'identity.json')) | ConvertFrom-Json
    $process=Get-Process -Id $identity.pid
    if ($process.Path -ne $ShellBinary -or $process.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $identity.filetime) { throw 'Fixture identity mismatch' }
    $result.identity=$identity
    if ($OutputTiming -eq 'BeforeAttach') {
        [IO.File]::WriteAllText((Join-Path $root 'gate.txt'), 'go')
        $until=[DateTime]::UtcNow.AddSeconds(10)
        while (!(Test-Path (Join-Path $root 'done.txt')) -and [DateTime]::UtcNow -lt $until) { Start-Sleep -Milliseconds 50 }
        if (!(Test-Path (Join-Path $root 'done.txt'))) { throw 'Fixture output not complete' }
    }
    $reg=$ns + '__' + $s.backend_id
    $port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg + '.port')))).Trim()
    $key=([IO.File]::ReadAllText((Join-Path $data ($reg + '.key')))).Trim()
    $tcp=[Net.Sockets.TcpClient]::new(); $tcp.Connect('127.0.0.1', $port)
    $stream=$tcp.GetStream(); $stream.ReadTimeout=15000; $stream.WriteTimeout=3000
    $reader=[IO.StreamReader]::new($stream, [Text.UTF8Encoding]::new($false), $false, 4096, $true)
    Send "AUTH $key"; if ($reader.ReadLine() -ne 'OK') { throw 'Authentication failed' }
    Send 'PERSISTENT'
    if ($OutputTiming -eq 'BeforeAttach') {
        $f=WaitFrame {param($x) ((Lines $x) -join "`n") -match 'HISTORY_END'}; $pane=$f.layout.id
        SaveStage 'detached-before' $f; Capture 'detached-before'
    }
    Send 'client-attach'; Send 'client-size 94 25'
    $f=WaitFrame {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25 -and ($OutputTiming -eq 'BeforeAttach' -or ((Lines $x) -join "`n") -match 'BOOT_READY')}
    $pane=$f.layout.id; [IO.File]::WriteAllText((Join-Path $root 'gate.txt'), 'go')
    $f=WaitFrame {param($x) ((Lines $x) -join "`n") -match 'HISTORY_END'}; SaveStage 'live-before' $f; Capture 'live-before'
    if (!$SkipEarlyCopy) {
        Send "copy-mode -t %$pane"; $f=WaitFrame {param($x) $x.layout.copy_mode}; SaveStage 'copy-before' $f
        Hex ('?' + ('x' * 80) + '中文tail')
        $f=WaitFrame {param($x) $x.status_message -like '*中文tail*'}; SaveStage 'prompt-before' $f
    }
    if (!$SkipSecondaryResize) {
        Send 'client-size 61 18'; $f=WaitFrame {param($x) $x.layout.cols -eq 61 -and $x.layout.rows -eq 18}
        SaveStage 'narrow' $f; Capture 'narrow'
    }
    if (!$SkipEarlyCopy) {
        Send "send-keys -t %$pane Escape"; $f=WaitFrame {param($x) $x.status_message -notlike '*中文tail*'}; SaveStage 'cancel-prompt' $f
    }
    if (!$SkipSecondaryResize) {
        Send 'client-size 94 25'; $f=WaitFrame {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25}
        SaveStage 'wide-again' $f; Capture 'wide-again'
    }
    if ($SkipEarlyCopy) { Send "copy-mode -t %$pane"; $f=WaitFrame {param($x) $x.layout.copy_mode} }
    Hex '?SCROLL_NEEDLE_中文_α'; $f=WaitFrame {param($x) $x.status_message -like '*SCROLL_NEEDLE*'}; SaveStage 'needle-prompt' $f
    Send "send-keys -t %$pane Enter"
    $f=WaitFrame {param($x) @((Lines $x) | Where-Object { $_ -like '*SCROLL_NEEDLE*' }).Count -gt 0}
    SaveStage 'search-top' $f; Capture 'search-top'
    $result.complete=(@($captures | Where-Object { !$_.complete }).Count -eq 0)
    $result.result=if ($result.complete) { 'complete-in-probe' } else { 'history-incomplete' }
} catch {
    $failure=$_.Exception.Message; $result.result='harness-failed'; $result.error=$failure
} finally {
    if ($tcp) { $tcp.Dispose() }
    if ($s) {
        if ($DirectBackend) { & $PsmuxBinary -L $ns kill-session -t $s.backend_id | Out-Null }
        else { & $ProductBinary @g stop $s.session_id | Out-Null }
        if ($LASTEXITCODE -ne 0) { $result.cleanup_error='Stop failed' }
    }
    # Start may fail after the child exists but before returning Session JSON.
    # Preserve the independent identity and fail cleanup closed in that case.
    $identityFile=Join-Path $root 'identity.json'
    if (!$identity -and (Test-Path $identityFile) -and (Get-Item $identityFile).Length -le 2KB) {
        $identity=[IO.File]::ReadAllText($identityFile) | ConvertFrom-Json
        if ($identity.path -ne $ShellBinary -or $identity.pid -le 0 -or $identity.filetime -le 0) { throw 'Invalid fixture cleanup identity' }
        $result.identity=$identity
    }
    $listed=(& $ProductBinary @g list --json | Out-String).Trim(); $result.cleanup_json=$listed
    if ($identity) {
        $process=Get-Process -Id $identity.pid -ErrorAction SilentlyContinue
        $result.original_ended=(!$process -or $process.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $identity.filetime -or $process.Path -ne $identity.path)
    }
    if ($DirectBackend -and $port) {
        $check=[Net.Sockets.TcpClient]::new()
        try { $check.Connect('127.0.0.1', $port); $result.direct_endpoint_closed=$false }
        catch { $result.direct_endpoint_closed=$true }
        finally { $check.Dispose() }
    }
    $stages.ToArray() | ConvertTo-Json -Depth 20 | Set-Content (Join-Path $root 'stages.json') -Encoding utf8
    $captures.ToArray() | ConvertTo-Json -Depth 10 | Set-Content (Join-Path $root 'captures.json') -Encoding utf8
    $result | ConvertTo-Json -Depth 10 | Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
$result | ConvertTo-Json -Depth 10
if ($failure -or $listed -ne '[]' -or $result.original_ended -ne $true -or $result.cleanup_error -or ($DirectBackend -and !$result.direct_endpoint_closed)) { throw 'Diagnostic or cleanup failed; inspect result.json' }
if (!$ObserveOnly -and !$result.complete) { throw 'Numbered output was lost or duplicated; inspect captures.json' }
