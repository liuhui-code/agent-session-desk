# Resource UI helpers; caller supplies the isolated process and mux scope.
function Start-Manager {
    param([object]$Case)
    $Case.ui_sequence++
    $Case.ui_name = 'ui' + $Case.ui_sequence
    $launch = @('new-session', '-d', '-s', $Case.ui_name, '-x', '120', '-y', '32', '--', $binaries.product, '--backend', $binaries.backend, '--runtime-dir', $Case.runtime, '--namespace', $Case.namespace)
    $null = Invoke-Mux $Case $launch -Driver
    $pane = (Invoke-Mux $Case @('list-panes', '-t', $Case.ui_name, '-F', '#{pane_id}') -Driver).Trim()
    if ($pane -notmatch '^%\d+$') { throw 'Driver must contain one known original pane.' }
    $Case.ui_target = $Case.ui_name + ':.' + $pane
    $warm = (Invoke-Mux $Case @('show-options', '-t', $Case.ui_name, '-g', '-v', 'warm') -Driver).Trim()
    $pool = (Invoke-Mux $Case @('show-options', '-t', $Case.ui_name, '-g', '-v', 'warm-pool-size') -Driver).Trim()
    if ($warm -ne 'off' -or $pool -ne '0') { throw "Driver warm configuration differs: warm=$warm pool=$pool" }
    $ready = [Diagnostics.Stopwatch]::StartNew()
    do {
        $screen = Invoke-Mux $Case @('capture-pane', '-t', $Case.ui_target, '-p') -Driver
        if (Test-AsdResourceUIReady $screen $Case.count) {
            if ($Case.count -gt 0) { Confirm-AsdResourceTabInput $Case }
            $screen | Set-Content -Encoding utf8 (Join-Path $evidence "$($Case.id)-ui-$($Case.ui_sequence).txt")
            $snapshot = Read-Snapshot $Case.id 'ui-ready' $Case.count
            $manager = @($snapshot.processes | Where-Object { $_.executable_path -eq $binaries.product })
            if ($manager.Count -ne 1) { throw 'TUI screen appeared without one live manager identity.' }
            $Case.ui_identity = $manager[0]
            Assert-AsdResourceDriverHost $Case $driverHost
            Write-Event 'ui-started' @{ scenario = $Case.id; identity = $Case.ui_identity; width = 120; height = 32 }
            return
        }
        Start-Sleep -Milliseconds 250
    } while ($ready.Elapsed.TotalSeconds -lt 20)
    $screen | Set-Content -Encoding utf8 (Join-Path $evidence "$($Case.id)-ui-failed.txt")
    throw 'Normal Tab workspace did not become ready in its driver ConPTY.'
}

function Get-AsdResourceUIFocus {
    param([string]$Screen)
    $rows = @($Screen -split "`r?`n" | Where-Object { $_.Trim().Length -gt 0 })
    if ($rows.Count -gt 0) {
        $footer = $rows[-1].Trim()
        if ($footer -match '^管理模式(?:\s|·|$)') { return 'management' }
        if ($footer -match '^(终端模式|滚动历史|滚动位置)(?:\s|·|$)') { return 'terminal' }
    }
    throw 'Resource UI focus is unknown; no exit text was sent.'
}

function Stop-Manager {
    param([object]$Case)
    $screen = Invoke-Mux $Case @('capture-pane', '-t', $Case.ui_target, '-p') -Driver
    $before = Get-AsdResourceUIFocus $screen
    if ($before -eq 'terminal') {
        $null = Invoke-Mux $Case @('send-keys', '-t', $Case.ui_target, 'F6') -Driver
        $wait = [Diagnostics.Stopwatch]::StartNew()
        do {
            $screen = Invoke-Mux $Case @('capture-pane', '-t', $Case.ui_target, '-p') -Driver
            $focus = 'unknown'
            try { $focus = Get-AsdResourceUIFocus $screen } catch { }
            if ($focus -eq 'management') { break }
            Start-Sleep -Milliseconds 100
        } while ($wait.Elapsed.TotalSeconds -lt 5)
        if ($focus -ne 'management') { throw 'Resource UI did not confirm management focus after F6; q was not sent.' }
    }
    Write-Event 'ui-management-focus-confirmed' @{ scenario = $Case.id; previous_focus = $before; screen = $screen }
    $null = Invoke-Mux $Case @('send-keys', '-t', $Case.ui_target, '-l', 'q') -Driver
    $key = "$($Case.ui_identity.process_id):$($Case.ui_identity.creation_filetime)"
    if (-not $script:held.ContainsKey($key)) { throw 'Missing retained TUI identity handle.' }
    if (-not $script:held[$key].process.WaitForExit(15000)) { throw 'TUI did not exit normally after management q.' }
    Write-Event 'ui-exited-after-q' @{ scenario = $Case.id; identity = $Case.ui_identity; exit_code = $script:held[$key].process.ExitCode }
    if ($script:held[$key].process.ExitCode -ne 0) { throw 'TUI exited with a failure status.' }
    $null = Invoke-Mux $Case @('kill-session', '-t', $Case.ui_name) -Driver
    $Case.ui_target = $null
}

function Test-AsdResourceUIReady {
    param([string]$Screen, [int]$Count)
    if ($Screen -notmatch 'Agent Session Desk' -or $Screen -match '正在连接原终端|终端不可交互|只读：') { return $false }
    try { $focus = Get-AsdResourceUIFocus $Screen } catch { return $false }
    if ($Count -eq 0) { return $focus -eq 'management' }
    return $focus -eq 'terminal' -and $Screen -match 'session-00'
}

function Confirm-AsdResourceTabInput {
    param([object]$Case)
    # A controlled fixture, not an agent-readiness inference. Submit exactly
    # once through the normal TUI path; never retry a missing acknowledgement.
    $marker = "resource-ui-ready-$($Case.ui_sequence)"
    $null = Invoke-Mux $Case @('send-keys', '-t', $Case.ui_target, '-l', $marker) -Driver
    $null = Invoke-Mux $Case @('send-keys', '-t', $Case.ui_target, 'Enter') -Driver
    $wait = [Diagnostics.Stopwatch]::StartNew()
    do {
        $observations = @(Get-AsdFixtureJournal $Case.fixtures[0].report | Where-Object { $_.event -eq 'input' -and $_.text -eq $marker })
        if ($observations.Count -gt 1) { throw 'TUI readiness marker was delivered more than once.' }
        if ($observations.Count -eq 1) {
            Write-Event 'ui-tab-input-confirmed' @{ scenario = $Case.id; marker = $marker; session_id = $Case.sessions[0].session_id; run_id = $Case.sessions[0].run_id; fixture_pid = $Case.fixtures[0].pid }
            return
        }
        Start-Sleep -Milliseconds 100
    } while ($wait.Elapsed.TotalSeconds -lt 10)
    throw 'Controlled fixture did not acknowledge the TUI marker; input was not retried.'
}
