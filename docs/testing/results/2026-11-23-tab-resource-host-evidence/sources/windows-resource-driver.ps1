# Measurement driver only. The product backend uses its own clean environment.
function Resolve-AsdResourceDriverHost {
    param([string]$Directory, [object]$Pin)
    if (-not $Directory) { return [pscustomobject]@{ mode = 'system'; directory = $null; files = @(); version = $null } }
    $directoryPath = [IO.Path]::GetFullPath($Directory)
    $files = @()
    foreach ($entry in $Pin.files.PSObject.Properties) {
        $path = Join-Path $directoryPath $entry.Name
        $item = Get-Item -LiteralPath $path
        if ($item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw 'Driver host must contain regular pinned files.' }
        $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($hash -cne $entry.Value) { throw 'Measurement host SHA-256 differs from the pinned candidate.' }
        $signature = Get-AuthenticodeSignature -LiteralPath $path
        if ($signature.Status -ne 'Valid' -or $signature.SignerCertificate.Subject -notmatch '^CN=Microsoft Corporation,') { throw 'Measurement host Microsoft signature is not valid.' }
        $files += [pscustomobject]@{ name = $entry.Name; path = $path; sha256 = $hash; signer = $signature.SignerCertificate.Subject }
    }
    return [pscustomobject]@{ mode = 'sdk-candidate'; directory = $directoryPath; files = $files; version = $Pin.version }
}

function Get-AsdResourceDriverEnvironment {
    param([string]$DataDirectory, [string]$Config, [object]$HostSelection)
    $environment = Get-MuxEnvironment $DataDirectory $Config
    if ($HostSelection.mode -eq 'sdk-candidate') { $environment['PSMUX_CONPTY_DIR'] = $HostSelection.directory }
    return $environment
}

function Assert-AsdResourceDriverHost {
    param([object]$Case, [object]$HostSelection)
    if ($HostSelection.mode -eq 'system') { return }
    $registry = @(Get-RegistryEvidence $Case.driver_data $binaries.driver | Where-Object live_identity_verified | Sort-Object process_id -Unique)
    $dll = @($HostSelection.files | Where-Object name -eq 'conpty.dll')[0]
    $executable = @($HostSelection.files | Where-Object name -eq 'OpenConsole.exe')[0]
    $modules = @(); $hosts = @()
    foreach ($entry in $registry) {
        $process = Get-Process -Id $entry.process_id
        try {
            if ($process.Path -ne $binaries.driver -or $process.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() -ne $entry.creation_filetime) { throw 'Measurement driver identity changed.' }
            $modules += @($process.Modules | Where-Object { (Get-AsdMetricPathKey $_.FileName) -ceq (Get-AsdMetricPathKey $dll.path) } | ForEach-Object {
                if ((Get-FileHash $_.FileName).Hash.ToLowerInvariant() -cne $dll.sha256) { throw 'Loaded driver DLL changed.' }
                [pscustomobject]@{ driver_pid = $process.Id; driver_filetime = $entry.creation_filetime; module = $_.FileName; sha256 = $dll.sha256 }
            })
        } finally { $process.Dispose() }
    }
    foreach ($candidate in @(Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -in @($registry.process_id) })) {
        $child = Get-Process -Id $candidate.ProcessId -ErrorAction SilentlyContinue
        if ($null -eq $child) { continue }
        try {
            if ((Get-AsdMetricPathKey $child.Path) -ceq (Get-AsdMetricPathKey $executable.path)) {
                if ((Get-FileHash $child.Path).Hash.ToLowerInvariant() -cne $executable.sha256) { throw 'Observed console host changed.' }
                $hosts += [pscustomobject]@{ process_id = $child.Id; creation_filetime = $child.StartTime.ToUniversalTime().ToFileTimeUtc().ToString(); executable_path = $child.Path; parent_process_id = $candidate.ParentProcessId; sha256 = $executable.sha256 }
            }
        } finally { $child.Dispose() }
    }
    Write-Event 'measurement-host-observed' @{ scenario = $Case.id; modules = $modules; hosts = $hosts }
    if ($modules.Count -ne 1 -or $hosts.Count -ne 1) { throw 'Pinned measurement DLL and host were not both observed; fallback is not accepted.' }
}
