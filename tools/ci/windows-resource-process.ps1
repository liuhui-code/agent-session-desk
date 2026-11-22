# Helpers extracted unchanged from windows-resources.ps1; caller owns scope.
function Write-Event {
    param([string]$Kind, [object]$Data)
    [ordered]@{ event = $Kind; utc = [DateTime]::UtcNow.ToString('o'); elapsed_ms = $clock.Elapsed.TotalMilliseconds; data = $Data } |
        ConvertTo-Json -Depth 20 -Compress | Add-Content -Encoding utf8 (Join-Path $evidence 'events.jsonl')
}

function Start-RecordedProcess {
    param([string]$Executable, [string[]]$ArgumentList, [hashtable]$Environment = @{}, [switch]$DiscardOutput)
    $info = [Diagnostics.ProcessStartInfo]::new()
    $info.FileName = $Executable
    $info.WorkingDirectory = $experiment
    $info.UseShellExecute = $false
    $info.CreateNoWindow = $true
    $info.RedirectStandardInput = $true
    $info.RedirectStandardOutput = $true
    $info.RedirectStandardError = $true
    foreach ($value in $ArgumentList) { $info.ArgumentList.Add($value) }
    foreach ($key in @($info.Environment.Keys)) {
        if ($key -like 'PSMUX_*' -or $key -in @('TMUX', 'TMUX_PANE', 'TMUX_TMPDIR') -or $key -like 'ASD_FIXTURE_*') {
            [void]$info.Environment.Remove($key)
        }
    }
    foreach ($key in $Environment.Keys) { $info.Environment[$key] = $Environment[$key] }
    $process = [Diagnostics.Process]::new()
    $process.StartInfo = $info
    if (-not $process.Start()) { throw "Could not start $Executable" }
    if ($DiscardOutput) {
        $stdout = $process.StandardOutput.BaseStream.CopyToAsync([IO.Stream]::Null)
        $stderr = $process.StandardError.BaseStream.CopyToAsync([IO.Stream]::Null)
    } else {
        $stdout = $process.StandardOutput.ReadToEndAsync()
        $stderr = $process.StandardError.ReadToEndAsync()
    }
    return [pscustomobject]@{ process = $process; stdout = $stdout; stderr = $stderr }
}

function Get-MuxEnvironment {
    param([string]$DataDirectory, [string]$Config)
    return @{ PSMUX_DATA_DIR = $DataDirectory; PSMUX_CONFIG_FILE = $Config; PSMUX_NO_WARM = '1'; PSMUX_ALLOW_NESTING = '1'; TERM = 'xterm-256color' }
}
