[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string]$PsmuxBinary,
    [string]$ProductBinary = 'build/asd-windows-native.exe',
    [string]$FixtureBinary = 'build/asd-fixture-windows-native.exe',
    [string]$EvidenceDirectory = 'build/ci-windows/resources',
    [ValidateRange(1, 10)][int]$WarmupSeconds = 3,
    [ValidateRange(8, 15)][int]$SampleSeconds = 12
)

# A bounded, native Windows baseline, not the 15-minute/long-soak catalog tests.
# The normal TUI runs inside a separate measurement-only psmux ConPTY. Its
# driver and sampler are reported separately, never silently charged to asd.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if (-not $IsWindows) { throw 'Resource evidence must be collected on native Windows.' }
if ($WarmupSeconds + $SampleSeconds -gt 20) { throw 'The 25-second fixture load needs at least five seconds of control/staggering margin.' }
. (Join-Path $PSScriptRoot 'windows-process-metrics.ps1')
. (Join-Path $PSScriptRoot 'windows-resource-process.ps1')
. (Join-Path $PSScriptRoot 'windows-resource-ui.ps1')
. (Join-Path $PSScriptRoot 'windows-fixture-journal.ps1')
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$evidence = [IO.Path]::GetFullPath($EvidenceDirectory)
$experiment = Join-Path $repo ('build/resource-runtime-' + [Guid]::NewGuid().ToString('N'))
$lock = Get-Content -Raw (Join-Path $PSScriptRoot 'windows-dependencies.json') | ConvertFrom-Json
$originals = @{
    backend = (Resolve-Path -LiteralPath $PsmuxBinary).Path
    driver = (Resolve-Path -LiteralPath $PsmuxBinary).Path
    product = (Resolve-Path -LiteralPath $ProductBinary).Path
    managed = (Resolve-Path -LiteralPath $FixtureBinary).Path
    direct = (Resolve-Path -LiteralPath $FixtureBinary).Path
}
if ((Get-FileHash -LiteralPath $originals.backend -Algorithm SHA256).Hash.ToLowerInvariant() -ne $lock.psmux.executable_sha256) {
    throw 'Resource baseline requires the pinned official psmux binary.'
}
if (Test-Path -LiteralPath (Join-Path $evidence 'summary.json')) { throw 'Choose a fresh evidence directory.' }
New-Item -ItemType Directory -Force $evidence, $experiment | Out-Null
$binaries = @{}
$hashes = [ordered]@{}
foreach ($key in $originals.Keys) {
    $folder = Join-Path $experiment "bin/$key"
    New-Item -ItemType Directory -Force $folder | Out-Null
    $leaf = switch ($key) { { $_ -in 'backend', 'driver' } { 'psmux.exe' } 'product' { 'asd.exe' } default { 'asd-fixture.exe' } }
    $binaries[$key] = Join-Path $folder $leaf
    Copy-Item -LiteralPath $originals[$key] -Destination $binaries[$key]
    $hashes[$key] = (Get-FileHash -LiteralPath $binaries[$key] -Algorithm SHA256).Hash.ToLowerInvariant()
}
$imageRoles = @{
    $binaries.backend = 'backend_client'; $binaries.driver = 'measurement_driver'
    $binaries.product = 'product_manager'; $binaries.managed = 'managed_fixture'
    $binaries.direct = 'direct_fixture'
}
$clock = [Diagnostics.Stopwatch]::StartNew()
$script:sampleIndex = 0
$script:previousOwned = @()
$script:held = @{}
$script:phases = [Collections.Generic.List[object]]::new()
$script:commands = [Collections.Generic.List[object]]::new()
$script:cases = [Collections.Generic.List[object]]::new()
$script:cleanup = [Collections.Generic.List[object]]::new()
$failure = $null


function Invoke-Control {
    param([string]$Label, [string]$Executable, [string[]]$ArgumentList, [hashtable]$Environment = @{})
    $timer = [Diagnostics.Stopwatch]::StartNew()
    $child = Start-RecordedProcess $Executable $ArgumentList $Environment
    try {
        if (-not $child.process.WaitForExit(45000)) {
            # This is the exact Process handle returned by our own Start, not a
            # later PID lookup or process-name kill. Do not kill unknown trees.
            $child.process.Kill()
            throw "$Label exceeded its 45-second control deadline"
        }
        $output = $child.stdout.GetAwaiter().GetResult()
        $errorText = $child.stderr.GetAwaiter().GetResult()
        $record = [ordered]@{
            label = $Label; elapsed_ms = $timer.Elapsed.TotalMilliseconds
            exit_code = $child.process.ExitCode; arguments = $ArgumentList
            stdout = $output; stderr = $errorText
        }
        $script:commands.Add($record)
        Write-Event 'control' $record
        if ($child.process.ExitCode -ne 0) { throw "$Label failed ($($child.process.ExitCode)): $errorText" }
        return $output
    } finally { $child.process.Dispose() }
}

function Invoke-Product {
    param([object]$Case, [string[]]$ArgumentList)
    return Invoke-Control ('asd-' + $ArgumentList[0]) $binaries.product (@('--backend', $binaries.backend, '--runtime-dir', $Case.runtime, '--namespace', $Case.namespace) + $ArgumentList)
}

function Invoke-Mux {
    param([object]$Case, [string[]]$ArgumentList, [switch]$Driver)
    if ($Driver) {
        return Invoke-Control 'driver' $binaries.driver (@('-L', $Case.driver_namespace, '-f', $Case.driver_config) + $ArgumentList) (Get-MuxEnvironment $Case.driver_data $Case.driver_config)
    }
    return Invoke-Control 'backend-control' $binaries.backend (@('-L', $Case.namespace, '-f', $Case.config) + $ArgumentList) (Get-MuxEnvironment $Case.data $Case.config)
}

function Wait-Journal {
    param([string]$Path, [string]$Event, [int]$LoadID = 0, [int]$TimeoutSeconds = 15)
    $wait = [Diagnostics.Stopwatch]::StartNew()
    do {
        $found = @(Get-AsdFixtureJournal $Path | Where-Object { $_.event -eq $Event -and ($LoadID -eq 0 -or $_.load_id -eq $LoadID) })
        if ($found.Count -gt 0) { return $found[-1] }
        Start-Sleep -Milliseconds 100
    } while ($wait.Elapsed.TotalSeconds -lt $TimeoutSeconds)
    throw "No $Event/$LoadID observation in fixture journal $Path"
}

function Remember-ProcessHandles {
    param([object[]]$Identities)
    foreach ($key in @($script:held.Keys)) {
        if ($script:held[$key].process.HasExited) {
            $script:held[$key].process.Dispose()
            $script:held.Remove($key)
        }
    }
    foreach ($identity in $Identities) {
        if ($identity.executable_path -notin $imageRoles.Keys) { continue }
        $key = "$($identity.process_id):$($identity.creation_filetime)"
        if ($script:held.ContainsKey($key)) { continue }
        $process = $null
        try {
            $process = Get-Process -Id $identity.process_id -ErrorAction Stop
            $null = $process.Handle # Keep the original kernel object, including across PID reuse.
            if ($process.StartTime.ToFileTimeUtc().ToString() -ne $identity.creation_filetime -or
                $process.MainModule.FileName -ne $identity.executable_path) { $process.Dispose(); continue }
            $script:held[$key] = [pscustomobject]@{ identity = $identity; process = $process }
        } catch { if ($null -ne $process) { $process.Dispose() } }
    }
}

function Read-Snapshot {
    param([string]$Scenario, [string]$Phase, [int]$Count)
    $collection = [Diagnostics.Stopwatch]::StartNew()
    $roots = @()
    $registry = @()
    foreach ($case in $script:cases) {
        if ($case.mode -ne 'managed') { continue }
        $registry += @(Get-RegistryEvidence $case.data $binaries.backend)
    }
    foreach ($identity in @($registry | Where-Object { $_.live_identity_verified } | Select-Object process_id, creation_filetime -Unique)) {
        $roots += [pscustomobject]@{
            process_id = $identity.process_id; creation_filetime = $identity.creation_filetime
            executable_path = $binaries.backend; role = 'backend_server'; context = 'backend_server'
        }
    }
    $snapshot = Get-AsdWindowsProcessSnapshot -ImageRoles $imageRoles -RootIdentities $roots -PreviousOwned $script:previousOwned -SampleIndex $script:sampleIndex -ElapsedMilliseconds $clock.Elapsed.TotalMilliseconds
    $script:sampleIndex++
    $script:previousOwned = @($snapshot.owned_identities)
    Remember-ProcessHandles $script:previousOwned
    $snapshot | Add-Member -NotePropertyName registry -NotePropertyValue $registry
    $snapshot | Add-Member -NotePropertyName collection_with_registry_ms -NotePropertyValue $collection.Elapsed.TotalMilliseconds
    $snapshot | Add-Member -NotePropertyName retained_owned_handles -NotePropertyValue $script:held.Count
    $record = [ordered]@{ scenario = $Scenario; phase = $Phase; session_count = $Count; snapshot = $snapshot }
    $record | ConvertTo-Json -Depth 20 -Compress | Add-Content -Encoding utf8 (Join-Path $evidence 'samples.jsonl')
    return $snapshot
}

function Measure-Phase {
    param([object]$Case, [string]$Phase, [int]$ExpectedManagers)
    Write-Host "Resource baseline: $($Case.id) / $Phase"
    Start-Sleep -Seconds $WarmupSeconds
    $first = $script:sampleIndex
    $timer = [Diagnostics.Stopwatch]::StartNew()
    do {
        $iteration = [Diagnostics.Stopwatch]::StartNew()
        $snapshot = Read-Snapshot $Case.id $Phase $Case.count
        $fixturePath = if ($Case.mode -eq 'direct') { $binaries.direct } else { $binaries.managed }
        $fixtureCount = @($snapshot.processes | Where-Object { $_.executable_path -eq $fixturePath }).Count
        $managerCount = @($snapshot.processes | Where-Object { $_.executable_path -eq $binaries.product }).Count
        $serverCount = @($snapshot.processes | Where-Object { $_.role -eq 'backend_server' -and -not $_.is_descendant }).Count
        $expectedServers = if ($Case.mode -eq 'managed') { $Case.count } else { 0 }
        if ($fixtureCount -ne $Case.count -or $managerCount -ne $ExpectedManagers) {
            throw "Unexpected measured identities in $($Case.id)/${Phase}: fixture=$fixtureCount manager=$managerCount"
        }
        if ($serverCount -ne $expectedServers) { throw "Expected $expectedServers live registry-verified backend servers, measured $serverCount." }
        $delay = [Math]::Max(0, 1000 - $iteration.Elapsed.TotalMilliseconds)
        if ($delay -gt 0) { Start-Sleep -Milliseconds ([int]$delay) }
    } while ($timer.Elapsed.TotalSeconds -lt $SampleSeconds)
    $record = [ordered]@{
        scenario = $Case.id; phase = $Phase; session_count = $Case.count
        first_sample = $first; sample_count = $script:sampleIndex - $first
        warmup_seconds = $WarmupSeconds; elapsed_ms = $timer.Elapsed.TotalMilliseconds
        expected_managers = $ExpectedManagers; selected_session = $Case.selected_session
    }
    $script:phases.Add($record)
    Write-Event 'phase-completed' $record
    if ($record.sample_count -lt 6) { throw 'Fewer than six samples; collection was too slow for this short window.' }
}

function Get-RegistryEvidence {
    param([string]$DataDirectory, [string]$ExpectedImage)
    $files = @(Get-ChildItem -LiteralPath $DataDirectory -Filter '*.pid' -File -ErrorAction SilentlyContinue)
    $markers = Join-Path $DataDirectory 'servers'
    if (Test-Path -LiteralPath $markers) { $files += @(Get-ChildItem -LiteralPath $markers -File) }
    $records = @()
    foreach ($file in $files) {
        # Stable psmux rewrites server markers every five seconds. A shared
        # read may see that bounded truncate/write interval; retry observation
        # only, never a process start or input write. Persistent failure is not
        # a silently absent server.
        $valid = $false
        for ($attempt = 1; $attempt -le 5; $attempt++) {
            try {
                $value = (Get-AsdSharedText $file.FullName).Trim()
                $valid = $value -match '^(\d+):(\d+)$'
                if ($valid) { break }
            } catch { if ($attempt -eq 5) { throw } }
            if ($attempt -lt 5) { Start-Sleep -Milliseconds 20 }
        }
        if (-not $valid) { throw "Invalid owned server identity marker after bounded rereads: $($file.Name)" }
        $processID = [int]$Matches[1]
        $created = [string]$Matches[2]
        $live = $false
        $process = $null
        try {
            $process = Get-Process -Id $processID -ErrorAction Stop
            $live = $process.StartTime.ToFileTimeUtc().ToString() -eq $created -and $process.MainModule.FileName -eq $ExpectedImage
            if ($live) { Remember-ProcessHandles @([pscustomobject]@{ process_id = $processID; creation_filetime = $created; executable_path = $ExpectedImage }) }
        } catch { $live = $false } finally { if ($null -ne $process) { $process.Dispose() } }
        $records += [pscustomobject][ordered]@{ marker = [IO.Path]::GetRelativePath($DataDirectory, $file.FullName); process_id = $processID; creation_filetime = $created; live_identity_verified = $live; read_attempts = $attempt }
    }
    return $records
}

function Check-Backend {
    param([object]$Case, [string]$Phase)
    $listed = @(Invoke-Product $Case @('list', '--json') | ConvertFrom-Json | ForEach-Object { $_ })
    if ($listed.Count -ne $Case.count) { throw 'Live session count changed.' }
    foreach ($item in $Case.sessions) {
        $match = @($listed | Where-Object { $_.session_id -eq $item.session_id -and $_.run_id -eq $item.run_id -and $_.terminal_state -eq 'available' })
        if ($match.Count -ne 1) { throw 'Session/Run identity or terminal availability changed.' }
        $warm = (Invoke-Mux $Case @('show-options', '-t', $item.backend_id, '-g', '-v', 'warm')).Trim()
        $pool = (Invoke-Mux $Case @('show-options', '-t', $item.backend_id, '-g', '-v', 'warm-pool-size')).Trim()
        if ($warm -ne 'off' -or $pool -ne '0') { throw "Unexpected warm configuration: warm=$warm pool=$pool" }
        Write-Event 'backend-options' @{ scenario = $Case.id; phase = $Phase; backend_id = $item.backend_id; warm = $warm; warm_pool_size = $pool }
    }
    $registry = @(Get-RegistryEvidence $Case.data $binaries.backend)
    $live = @($registry | Where-Object { $_.live_identity_verified } | Select-Object process_id, creation_filetime -Unique)
    if ($live.Count -ne $Case.count) { throw "Expected $($Case.count) backend servers; verified $($live.Count)." }
    Write-Event 'backend-registry' @{ scenario = $Case.id; phase = $Phase; records = $registry }
}

function Start-Load {
    param([object]$Case, [object[]]$Fixtures)
    $loads = @()
    foreach ($fixture in $Fixtures) {
        $existing = @(Get-AsdFixtureJournal $fixture.report | Where-Object { $_.event -eq 'load-start' }).Count
        $command = 'load 100 256 25000'
        if ($Case.mode -eq 'direct') { $fixture.child.process.StandardInput.WriteLine($command); $fixture.child.process.StandardInput.Flush() }
        else {
            $null = Invoke-Mux $Case @('send-keys', '-t', $fixture.target, '-l', $command)
            $null = Invoke-Mux $Case @('send-keys', '-t', $fixture.target, 'Enter')
        }
        $start = Wait-Journal $fixture.report 'load-start' ($existing + 1)
        $loads += [pscustomobject]@{ fixture = $fixture; load_id = $start.load_id; started = $start }
    }
    return $loads
}

function Complete-Load {
    param([object]$Case, [object[]]$Loads, [string]$Phase)
    # Require every generator to cover the full sampling window. Control
    # latency starting other fixtures is visible, never hidden in nominal rate.
    $windowEnd = [DateTime]::UtcNow
    foreach ($load in $Loads) {
        $done = Wait-Journal $load.fixture.report 'load-finished' $load.load_id 35
        if ($done.status -ne 'completed' -or $done.lines_written -lt 2000 -or $done.bytes_written -ne $done.lines_written * 256) {
            throw 'Output generator did not complete its bounded declared workload.'
        }
        if ((Get-AsdFixtureFinishedUtc $done) -lt $windowEnd) { throw 'Output ended before sampling window completed.' }
        Write-Event 'load-observed' @{ scenario = $Case.id; phase = $Phase; report = [IO.Path]::GetFileName($load.fixture.report); start = $load.started; finish = $done }
        # The report is written before stdout under the same fixture lock. A
        # subsequent input observation proves it passed that completed write.
        $marker = "after-load-$($load.load_id)"
        if ($Case.mode -eq 'direct') { $load.fixture.child.process.StandardInput.WriteLine($marker); $load.fixture.child.process.StandardInput.Flush() }
        else {
            $null = Invoke-Mux $Case @('send-keys', '-t', $load.fixture.target, '-l', $marker)
            $null = Invoke-Mux $Case @('send-keys', '-t', $load.fixture.target, 'Enter')
        }
        $wait = [Diagnostics.Stopwatch]::StartNew()
        do {
            $observed = @(Get-AsdFixtureJournal $load.fixture.report | Where-Object { $_.event -eq 'input' -and $_.text -eq $marker }).Count -eq 1
            if ($observed) { break }
            Start-Sleep -Milliseconds 100
        } while ($wait.Elapsed.TotalSeconds -lt 10)
        if (-not $observed) { throw 'Fixture did not observe its post-load input marker.' }
    }
}

function New-Case {
    param([string]$Mode, [int]$Count)
    $id = "$Mode-$Count"
    $root = Join-Path $experiment $id
    $ns = 'asd-res-' + [Guid]::NewGuid().ToString('N').Substring(0, 12)
    $runtime = Join-Path $root 'runtime'
    $data = Join-Path $runtime $ns
    $driver = Join-Path $root 'driver'
    New-Item -ItemType Directory -Force $data, $driver | Out-Null
    $config = Join-Path $driver 'driver.conf'
    "set -g status off`nset -g history-limit 2000`nset-window-option -g remain-on-exit on`nset -g warm off`n" | Set-Content -Encoding utf8 $config
    return [pscustomobject]@{
        id = $id; mode = $Mode; count = $Count; namespace = $ns; runtime = $runtime; data = $data
        config = (Join-Path $data 'psmux-m0.conf'); driver_data = $driver; driver_config = $config
        driver_namespace = ('driver-' + $ns); sessions = @(); fixtures = @(); selected_session = $null
        ui_sequence = 0; ui_name = $null; ui_target = $null; ui_identity = $null
    }
}

function Close-Case {
    param([object]$Case)
    if ($Case.mode -eq 'managed') {
        if ($null -ne $Case.ui_target) { Stop-Manager $Case }
        foreach ($item in $Case.sessions) { $null = Invoke-Product $Case @('stop', $item.session_id) }
        $remaining = @(Invoke-Product $Case @('list', '--json') | ConvertFrom-Json | ForEach-Object { $_ })
        if ($remaining.Count -ne 0) { throw 'Product stop left discoverable sessions.' }
    } else {
        foreach ($fixture in $Case.fixtures) {
            $fixture.child.process.StandardInput.WriteLine('exit')
            $fixture.child.process.StandardInput.Flush()
            if (-not $fixture.child.process.WaitForExit(10000) -or $fixture.child.process.ExitCode -ne 0) { throw 'Direct fixture did not exit normally.' }
        }
    }
    $wait = [Diagnostics.Stopwatch]::StartNew()
    do {
        $snapshot = Read-Snapshot $Case.id 'cleanup' 0
        $remaining = @($snapshot.processes | Where-Object { $_.owner_role -in @($imageRoles.Values) -or $_.owner_role -eq 'backend_server' })
        $uncertain = @($snapshot.gaps | Where-Object { $_.reason -ne 'previous_owned_not_observed' })
        if ($remaining.Count -eq 0 -and $uncertain.Count -eq 0) { break }
        Start-Sleep -Milliseconds 200
    } while ($wait.Elapsed.TotalSeconds -lt 15)
    $record = [ordered]@{ scenario = $Case.id; normal_cleanup = ($remaining.Count -eq 0 -and $uncertain.Count -eq 0); remaining_owned_processes = $remaining; collection_gaps = $uncertain; elapsed_ms = $wait.Elapsed.TotalMilliseconds }
    $script:cleanup.Add($record)
    Write-Event 'cleanup' $record
    if (-not $record.normal_cleanup) { throw 'Owned process cleanup left live identities or an unresolved collection gap.' }
    foreach ($fixture in $Case.fixtures) {
        if ($Case.mode -eq 'direct') {
            $fixture.child.stdout.GetAwaiter().GetResult()
            $fixture.child.stderr.GetAwaiter().GetResult()
            $fixture.child.process.Dispose()
        }
    }
    foreach ($entry in $script:held.Values) { $entry.process.Dispose() }
    $script:held.Clear()
    $script:previousOwned = @()
}

try {
    $count = 1
        $case = New-Case 'managed' $count
        $script:cases.Add($case)
        for ($index = 0; $index -lt $count; $index++) {
            $report = Join-Path $evidence "$($case.id)-fixture-$index.jsonl"
            $created = Invoke-Product $case @('start', '--name', ('session-{0:D2}' -f $index), '--cwd', $experiment, '--env', "ASD_FIXTURE_REPORT_PATH=$report", '--env', 'ASD_FIXTURE_ENABLE_LOAD=1', '--', $binaries.managed) | ConvertFrom-Json
            $case.sessions += $created
            $started = Wait-Journal $report 'started'
            $pane = (Invoke-Mux $case @('list-panes', '-t', $created.backend_id, '-F', '#{pane_id}')).Trim()
            if ($pane -notmatch '^%\d+$') { throw 'Expected exactly one fixture pane.' }
            $case.fixtures += [pscustomobject]@{ report = $report; target = ($created.backend_id + ':.' + $pane); pid = $started.pid }
        }
        if ($count -gt 0) { $case.selected_session = $case.sessions[0].session_id }
        Check-Backend $case 'initial'
        if (Test-Path -LiteralPath $case.config) { Copy-Item -LiteralPath $case.config -Destination (Join-Path $evidence "$($case.id)-backend.conf") }
        Copy-Item -LiteralPath $case.driver_config -Destination (Join-Path $evidence "$($case.id)-driver.conf")
        Start-Manager $case

    $null = Invoke-Mux $case @('send-keys', '-t', $case.ui_target, 'f6') -Driver
    Start-Sleep -Milliseconds 700
    $screen = Invoke-Mux $case @('capture-pane', '-t', $case.ui_target, '-p') -Driver
    $focus = Get-AsdResourceUIFocus $screen
    Write-Event 'probe-lower-f6' @{ after = $screen; focus = $focus }
    if ($focus -ne 'management') {
        $registry = @(Get-RegistryEvidence $case.driver_data $binaries.driver | Where-Object live_identity_verified)
        if ($registry.Count -lt 1) { throw 'No live owned driver server for protocol diagnostic.' }
        $endpoint = Join-Path $case.driver_data ($case.driver_namespace + '__' + $case.ui_name)
        $port = [int](Get-AsdSharedText ($endpoint + '.port')).Trim()
        $secret = (Get-AsdSharedText ($endpoint + '.key')).Trim()
        $client = [Net.Sockets.TcpClient]::new(); $client.Connect('127.0.0.1', $port)
        $stream = $client.GetStream(); $stream.ReadTimeout = 2000
        $writer = [IO.StreamWriter]::new($stream, [Text.UTF8Encoding]::new($false), 1024, $true)
        $reader = [IO.StreamReader]::new($stream, [Text.Encoding]::UTF8, $false, 1024, $true)
        try {
            $writer.NewLine = "`n"; $writer.WriteLine('AUTH ' + $secret); $writer.Flush()
            if ($reader.ReadLine() -ne 'OK') { throw 'Driver auth failed; no key was sent.' }
            $writer.WriteLine('send-key f6'); $writer.Flush()
            Start-Sleep -Milliseconds 700
        } finally { $secret = $null; $writer.Dispose(); $reader.Dispose(); $stream.Dispose(); $client.Dispose() }
        $screen = Invoke-Mux $case @('capture-pane', '-t', $case.ui_target, '-p') -Driver
        $focus = Get-AsdResourceUIFocus $screen
        Write-Event 'probe-protocol-f6' @{ after = $screen; focus = $focus }
    }
    if ($focus -ne 'management') { throw 'Lowercase and direct protocol F6 did not confirm management focus.' }
    Stop-Manager $case
    Close-Case $case
} catch {
    $failure = $_.Exception.Message
    Write-Host "Resource baseline failed: $failure"
    Write-Event 'failure' @{ message = $failure; script_stack = $_.ScriptStackTrace }
} finally {
    # Failure-only backstop. Normal cleanup above is measured separately. Hold
    # and verify exact owned images; never kill by name, wildcard or PID alone.
    try { $null = Read-Snapshot 'experiment' 'final' 0 } catch {
        Write-Event 'final-snapshot-failed' @{ message = $_.Exception.Message }
        if ($null -eq $failure) { $failure = 'The final resource snapshot failed.' }
    }
    foreach ($entry in $script:held.Values) {
        try {
            if (-not $entry.process.HasExited) {
                $entry.process.Kill()
                $exited = $entry.process.WaitForExit(10000)
                Write-Event 'forced-owned-cleanup' @{ identity = $entry.identity; exited = $exited }
                if ($null -eq $failure) { $failure = 'Forced cleanup was required after the normal scenarios.' }
            }
        } catch {
            Write-Event 'forced-cleanup-failed' @{ identity = $entry.identity; message = $_.Exception.Message }
            if ($null -eq $failure) { $failure = 'Owned-process fallback cleanup failed.' }
        }
        finally { $entry.process.Dispose() }
    }
    $scenarioRecords = @($script:cases | ForEach-Object {
        [ordered]@{ id = $_.id; mode = $_.mode; count = $_.count; namespace = $_.namespace; sessions = $_.sessions; selected_session = $_.selected_session; fixture_reports = @($_.fixtures | ForEach-Object { [IO.Path]::GetFileName($_.report) }) }
    })
    $os = Get-CimInstance Win32_OperatingSystem
    $summary = [ordered]@{
        schema_version = 1; commit = (& git -C $repo rev-parse HEAD); completed_utc = [DateTime]::UtcNow.ToString('o')
        result = $(if ($null -eq $failure) { 'pass' } else { 'fail' }); failure = $failure
        elapsed_ms = $clock.Elapsed.TotalMilliseconds; sample_count = $script:sampleIndex
        windows_caption = $os.Caption; windows_version = $os.Version; image_os = $env:ImageOS; image_version = $env:ImageVersion
        total_visible_memory_kib = $os.TotalVisibleMemorySize; whole_machine_counters_collected = $false
        logical_processors = [Environment]::ProcessorCount; powershell = $PSVersionTable.PSVersion.ToString(); binary_sha256 = $hashes
        warmup_seconds = $WarmupSeconds; sample_seconds = $SampleSeconds; target_interval_ms = 1000
        load = @{ lines_per_second = 100; line_bytes = 256; duration_ms = 25000; output_only_stdout = $true }
        terminal = @{ width = 120; height = 32; history_limit = 2000; host = 'separate measurement-only psmux ConPTY'; interactive_windows_terminal = $false }
        scenarios = $scenarioRecords; phases = $script:phases.ToArray(); cleanup = $script:cleanup.ToArray()
        limitations = @(
            'Short baseline only: no performance budget, long-soak, leak-free, real-agent or Windows desktop acceptance.'
            'One-second live-process snapshots miss short-lived clients; observed CPU is a lower bound, not complete process-lifetime CPU.'
            'Working-set sums may double-count shared pages; private bytes are private committed memory, not physical residency.'
            'Driver/sampler resources are separate; shared WMI-provider overhead and driver interference cannot be subtracted away.'
            'Direct fixtures write to drained pipes; their output path is not equivalent to a terminal/ConPTY baseline.'
            'An unobserved or unassignable console helper is a coverage gap, not proof of zero auxiliary processes.'
            'Retained handles protect owned-image cleanup from PID reuse; exited handles are released at the next observation. This scaffolding can affect kernel/sampler cost.'
        )
    }
    $summary | ConvertTo-Json -Depth 24 | Set-Content -Encoding utf8 (Join-Path $evidence 'summary.json')
}
if ($null -ne $failure) { throw $failure }
