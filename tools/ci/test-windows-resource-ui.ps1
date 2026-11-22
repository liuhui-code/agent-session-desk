[CmdletBinding()]
param()
# Pure command-routing regression; real Tab/ConPTY/resource evidence is separate.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'windows-resource-ui.ps1')
$script:sent = [Collections.Generic.List[string]]::new()
$script:childInput = [Collections.Generic.List[string]]::new()
$script:focus = 'terminal'
$process = [pscustomobject]@{ exited = $false; ExitCode = 0 }
$process | Add-Member -MemberType ScriptMethod -Name WaitForExit -Value { param($Milliseconds) return $this.exited }
$script:held = @{ '42:100' = [pscustomobject]@{ process = $process } }
$case = [pscustomobject]@{ id = 'managed-1'; ui_target = 'ui1:.%0'; ui_name = 'ui1'; ui_identity = [pscustomobject]@{ process_id = 42; creation_filetime = '100' } }
function Invoke-Mux {
    param($Case, [string[]]$ArgumentList, [switch]$Driver)
    if (-not $Driver) { throw 'UI control escaped its measurement driver.' }
    if ($ArgumentList[0] -eq 'capture-pane') {
        $footer = switch ($script:focus) { 'terminal' { '终端模式 · F6 管理' } 'management' { '管理模式 · q 退出' } default { 'unknown owner' } }
        return "Agent Session Desk`nsession-00`n$footer"
    }
    if ($ArgumentList[0] -eq 'send-keys') {
        $key = $ArgumentList[-1]; $script:sent.Add($key)
        if ($key -eq 'F6') { $script:focus = 'management' }
        elseif ($script:focus -eq 'terminal') { $script:childInput.Add($key) }
        elseif ($key -eq 'q') { $script:held['42:100'].process.exited = $true }
        return ''
    }
    if ($ArgumentList[0] -eq 'kill-session') { return '' }
    throw 'Unexpected driver operation.'
}
function Write-Event { param($Kind, $Data) }
Stop-Manager $case
if (($script:sent -join ',') -ne 'F6,q' -or $script:childInput.Count -ne 0) { throw 'Exit keys reached the child terminal or skipped confirmed management focus.' }
$script:sent.Clear(); $script:childInput.Clear(); $script:focus = 'management'
$process.exited = $false; $case.ui_target = 'ui2:.%0'; $case.ui_name = 'ui2'
Stop-Manager $case
if (($script:sent -join ',') -ne 'q') { throw 'Management focus was toggled back into terminal before exit.' }
$script:sent.Clear(); $script:focus = 'unknown'; $case.ui_target = 'ui3:.%0'
$rejected = $false
try { Stop-Manager $case } catch { $rejected = $_.Exception.Message -eq 'Resource UI focus is unknown; no exit text was sent.' }
if (-not $rejected -or $script:sent.Count -ne 0) { throw 'Unknown focus allowed exit input.' }
foreach ($footer in @('终端模式 · F6 管理', '滚动历史 · 12 行', '滚动历史 · (search down) term', '滚动位置 · 3 行')) {
    if ((Get-AsdResourceUIFocus "header`n$footer") -ne 'terminal') { throw 'Terminal/history focus was not recognized.' }
}
$rejected = $false
try { $null = Get-AsdResourceUIFocus "管理模式 in child output`nunknown owner" } catch { $rejected = $true }
if (-not $rejected) { throw 'A mode word in child output was mistaken for the UI footer.' }
'Resource UI stop routing and footer guards passed. Pure simulated driver only; no native counters or TUI executed.'
if (Test-AsdResourceUIReady "Agent Session Desk`nsession-00`n正在连接原终端…`n终端模式 · F6 管理" 1) { throw 'A connecting Tab was accepted for resource sampling.' }
if (-not (Test-AsdResourceUIReady "Agent Session Desk`nsession-00`nfixture data`n终端模式 · F6 管理" 1)) { throw 'A rendered normal Tab was not accepted.' }
if (-not (Test-AsdResourceUIReady "Agent Session Desk`n暂无打开的 Tab`n管理模式 · q 退出" 0)) { throw 'Empty management workspace was not accepted.' }
if (Test-AsdResourceUIReady "Agent Session Desk`n暂无打开的 Tab`n终端模式 · F6 管理" 0) { throw 'Empty workspace still in terminal focus was accepted.' }
'Resource UI ready contracts passed; fixture input acknowledgement is separately checked in native execution.'
