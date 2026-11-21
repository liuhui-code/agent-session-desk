# Dot-source this helper. Only Get-AsdWindowsProcessSnapshot requires Windows;
# Resolve-AsdProcessOwnership is a pure graph operation for synthetic validation.
# ImageRoles must contain isolated, caller-owned full executable paths, never
# global names such as pwsh.exe. Explicit identities override image-path roles.
#
# Counter definitions and identity limits:
# https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.totalprocessortime
# https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.privatememorysize64
# https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.workingset64
# https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-process
# https://learn.microsoft.com/en-us/windows/win32/wmisdk/cim-datetime
# https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getprocesstimes
# Working sets include shared pages; their sum is not unique physical memory.
# CPU is cumulative user+kernel milliseconds, not a CPU percentage. Sampling
# misses processes that begin and end between snapshots: CPU deltas are a lower
# bound. The sampler group excludes shared WMI-provider work outside its tree.

function Get-AsdMetricProperty {
    param([object]$Object, [string]$Name, [object]$Default = $null)
    if ($Object -is [System.Collections.IDictionary]) {
        if ($Object.Contains($Name)) { return $Object[$Name] }
    } elseif ($null -ne $Object -and $null -ne $Object.PSObject.Properties[$Name]) {
        return $Object.PSObject.Properties[$Name].Value
    }
    return $Default
}

function Get-AsdMetricPathKey {
    param([Parameter(Mandatory = $true)][string]$Path)
    # A lexical Windows path key also works in Linux-only synthetic graph tests.
    # The live collector requires full paths and compares the actual image path.
    return $Path.Replace('/', '\').ToLowerInvariant()
}

function Test-AsdMetricIdentity {
    param([object]$Process, [object]$Identity)
    return (
        [int]$Process.process_id -eq [int]$Identity.process_id -and
        [string]$Process.creation_filetime -ceq [string]$Identity.creation_filetime -and
        (Get-AsdMetricPathKey $Process.executable_path) -ceq (Get-AsdMetricPathKey $Identity.executable_path)
    )
}

function Resolve-AsdProcessOwnership {
    [CmdletBinding()]
    param(
        [AllowEmptyCollection()][object[]]$Processes = @(),
        [System.Collections.IDictionary]$ImageRoles = @{},
        [AllowEmptyCollection()][object[]]$RootIdentities = @(),
        [AllowEmptyCollection()][object[]]$PreviousOwned = @()
    )
    Set-StrictMode -Version Latest
    $rows = @{}
    $anchors = @{}
    $previousDescendants = @{}
    $imageMap = @{}
    $gaps = [Collections.Generic.List[object]]::new()
    foreach ($key in $ImageRoles.Keys) {
        $imageMap[(Get-AsdMetricPathKey ([string]$key))] = [string]$ImageRoles[$key]
    }
    foreach ($row in $Processes) {
        $key = [string][int]$row.process_id
        if ($rows.ContainsKey($key)) { throw "Duplicate candidate process ID: $key" }
        if ($row.creation_filetime -isnot [string] -or $row.creation_filetime -notmatch '^[0-9]+$') {
            throw 'Creation FILETIME identities must be decimal strings.'
        }
        $null = [long]::Parse($row.creation_filetime, [Globalization.CultureInfo]::InvariantCulture)
        $rows[$key] = $row
        $imageKey = Get-AsdMetricPathKey $row.executable_path
        if ($imageMap.ContainsKey($imageKey)) {
            $anchors[$key] = [pscustomobject]@{
                role = $imageMap[$imageKey]; owner_role = $imageMap[$imageKey]
                context = $imageMap[$imageKey]; ownership_kind = 'image_path'; is_descendant = $false
                root_process_id = [int]$row.process_id; root_creation_filetime = $row.creation_filetime
            }
        }
    }
    foreach ($identity in $PreviousOwned) {
        $key = [string][int]$identity.process_id
        if (-not $rows.ContainsKey($key)) { continue }
        if (-not (Test-AsdMetricIdentity $rows[$key] $identity)) {
            $gaps.Add([pscustomobject]@{
                reason = 'previous_owned_identity_changed'; process_id = [int]$identity.process_id
                creation_filetime = $identity.creation_filetime
                owner_role = [string](Get-AsdMetricProperty $identity 'owner_role' $identity.role)
            })
            continue
        }
        if (-not $anchors.ContainsKey($key)) {
            $remembered = [pscustomobject]@{
                role = [string]$identity.role
                owner_role = [string](Get-AsdMetricProperty $identity 'owner_role' $identity.role)
                context = [string](Get-AsdMetricProperty $identity 'context' $identity.role)
                ownership_kind = 'previous_identity'
                is_descendant = [bool](Get-AsdMetricProperty $identity 'is_descendant' $true)
                root_process_id = [int](Get-AsdMetricProperty $identity 'root_process_id' $identity.process_id)
                root_creation_filetime = [string](Get-AsdMetricProperty $identity 'root_creation_filetime' $identity.creation_filetime)
            }
            if ($remembered.is_descendant) { $previousDescendants[$key] = $remembered }
            else { $anchors[$key] = $remembered }
        }
    }
    foreach ($identity in $RootIdentities) {
        $key = [string][int]$identity.process_id
        if ($identity.creation_filetime -isnot [string] -or $identity.creation_filetime -notmatch '^[0-9]+$') {
            throw 'Explicit root FILETIME identities must be decimal strings.'
        }
        if (-not $rows.ContainsKey($key) -or -not (Test-AsdMetricIdentity $rows[$key] $identity)) {
            $gaps.Add([pscustomobject]@{
                reason = 'explicit_root_missing_or_identity_changed'; process_id = [int]$identity.process_id
                creation_filetime = $identity.creation_filetime; owner_role = [string]$identity.role
            })
            continue
        }
        $anchors[$key] = [pscustomobject]@{
            role = [string]$identity.role; owner_role = [string]$identity.role
            context = [string](Get-AsdMetricProperty $identity 'context' $identity.role)
            ownership_kind = 'explicit_identity'; is_descendant = $false
            root_process_id = [int]$identity.process_id; root_creation_filetime = $identity.creation_filetime
        }
    }

    $owned = [Collections.Generic.List[object]]::new()
    foreach ($row in $Processes) {
        $cursor = $row
        $visited = [Collections.Generic.HashSet[int]]::new()
        $anchor = $null
        $fallback = $null
        $fallbackCursor = $null
        $reason = $null
        while ($null -ne $cursor) {
            $cursorKey = [string][int]$cursor.process_id
            if (-not $visited.Add([int]$cursor.process_id)) { $reason = 'ancestry_cycle'; break }
            if ($anchors.ContainsKey($cursorKey)) { $anchor = $anchors[$cursorKey]; break }
            # A remembered descendant can survive its parent exiting. It must
            # not mask a current root's newer role/context (for example, a
            # backend image later proven by the registry to be the server).
            if ($null -eq $fallback -and $previousDescendants.ContainsKey($cursorKey)) {
                $fallback = $previousDescendants[$cursorKey]
                $fallbackCursor = $cursor
            }
            $parentKey = [string][int]$cursor.parent_process_id
            if (-not $rows.ContainsKey($parentKey)) { $reason = 'parent_not_in_verified_snapshot'; break }
            $parent = $rows[$parentKey]
            if ([long]$parent.creation_filetime -gt [long]$cursor.creation_filetime) {
                $reason = 'parent_created_after_child'; break
            }
            $cursor = $parent
        }
        if ($null -eq $anchor -and $null -ne $fallback) {
            $anchor = $fallback
            $cursor = $fallbackCursor
        }
        if ($null -eq $anchor) {
            $gaps.Add([pscustomobject]@{
                reason = $reason; process_id = [int]$row.process_id
                creation_filetime = $row.creation_filetime; parent_process_id = [int]$row.parent_process_id
                owner_role = $null
            })
            continue
        }
        $isAnchor = [int]$cursor.process_id -eq [int]$row.process_id
        $isDescendant = -not $isAnchor -or $anchor.is_descendant
        $role = $(if ($isAnchor) { $anchor.role } else { $anchor.owner_role })
        # A console basename only classifies an already proven descendant; it
        # never makes a process globally owned. Keep the nearest root's role.
        $leaf = ($row.executable_path -replace '/', '\').Split('\')[-1]
        if ($isDescendant -and $leaf -in @('conhost.exe', 'OpenConsole.exe')) { $role = 'console_aux' }
        $owned.Add([pscustomobject]@{
            process_id = [int]$row.process_id; creation_filetime = $row.creation_filetime
            executable_path = $row.executable_path; role = $role; owner_role = $anchor.owner_role
            context = $anchor.context; is_descendant = $isDescendant
            ownership_kind = $(if ($isAnchor) { $anchor.ownership_kind } else { 'verified_current_ancestry' })
            root_process_id = $anchor.root_process_id; root_creation_filetime = $anchor.root_creation_filetime
        })
    }
    return [pscustomobject]@{ owned_identities = @($owned.ToArray()); gaps = @($gaps.ToArray()) }
}

function Get-AsdWindowsProcessSnapshot {
    [CmdletBinding()]
    param(
        [System.Collections.IDictionary]$ImageRoles = @{},
        [AllowEmptyCollection()][object[]]$RootIdentities = @(),
        [AllowEmptyCollection()][object[]]$PreviousOwned = @(),
        [int]$SampleIndex = 0,
        [double]$ElapsedMilliseconds = 0
    )
    Set-StrictMode -Version Latest
    if (-not $IsWindows) { throw 'Windows process counters require a native Windows host.' }
    $clock = [Diagnostics.Stopwatch]::StartNew()
    $sampleUtc = [DateTime]::UtcNow.ToString('o')
    $gaps = [Collections.Generic.List[object]]::new()
    $normalizedImages = @{}
    foreach ($key in $ImageRoles.Keys) {
        if (-not [IO.Path]::IsPathFullyQualified([string]$key)) { throw 'ImageRoles requires full executable paths.' }
        $normalizedImages[[IO.Path]::GetFullPath([string]$key)] = [string]$ImageRoles[$key]
    }
    $sampler = [Diagnostics.Process]::GetCurrentProcess()
    try {
        $allRoots = @($RootIdentities) + @([pscustomobject]@{
            process_id = $PID; creation_filetime = $sampler.StartTime.ToFileTimeUtc().ToString([Globalization.CultureInfo]::InvariantCulture)
            executable_path = $sampler.MainModule.FileName; role = 'sampler'; context = 'sampler'
        })
    } finally { $sampler.Dispose() }

    $cimRows = @()
    $cimClock = [Diagnostics.Stopwatch]::StartNew()
    try {
        # Do not fetch CommandLine or environment/owner information.
        $cimRows = @(Get-CimInstance -ClassName Win32_Process -Property ProcessId, ParentProcessId, CreationDate, Name, ExecutablePath -ErrorAction Stop)
    } catch {
        $gaps.Add([pscustomobject]@{
            reason = 'cim_snapshot_failed'; process_id = $null; owner_role = $null
            error_type = $_.Exception.GetType().FullName; error_hresult = $_.Exception.HResult
        })
    } finally { $cimClock.Stop() }
    $cimById = @{}
    $candidateIds = [Collections.Generic.HashSet[int]]::new()
    $imageKeys = @{}
    foreach ($key in $normalizedImages.Keys) { $imageKeys[(Get-AsdMetricPathKey $key)] = $normalizedImages[$key] }
    foreach ($row in $cimRows) {
        $key = [string][int]$row.ProcessId
        $cimById[$key] = $row
        if (-not [string]::IsNullOrEmpty([string]$row.ExecutablePath) -and
            $imageKeys.ContainsKey((Get-AsdMetricPathKey $row.ExecutablePath))) {
            $null = $candidateIds.Add([int]$row.ProcessId)
        }
    }
    foreach ($identity in @($allRoots) + @($PreviousOwned)) {
        $key = [string][int]$identity.process_id
        if ($cimById.ContainsKey($key)) { $null = $candidateIds.Add([int]$identity.process_id) }
        elseif ($PreviousOwned -contains $identity) {
            $gaps.Add([pscustomobject]@{
                reason = 'previous_owned_not_observed'; process_id = [int]$identity.process_id
                creation_filetime = $identity.creation_filetime; owner_role = $identity.owner_role
            })
        }
    }
    # This closure only chooses what to read. It does not establish ownership.
    do {
        $changed = $false
        foreach ($row in $cimRows) {
            if ($candidateIds.Contains([int]$row.ParentProcessId) -and $candidateIds.Add([int]$row.ProcessId)) { $changed = $true }
        }
    } while ($changed)

    $verified = [Collections.Generic.List[object]]::new()
    # Read sampler last so its observation includes most of this sample's work.
    $candidateOrder = @($candidateIds | Where-Object { $_ -ne $PID } | Sort-Object) + @($PID)
    foreach ($candidateId in $candidateOrder) {
        $key = [string][int]$candidateId
        if (-not $cimById.ContainsKey($key)) { continue }
        $cim = $cimById[$key]
        $process = $null
        $check = $null
        $path = $null
        $processName = $null
        $readReason = 'process_read_failed'
        try {
            $process = Get-Process -Id $candidateId -ErrorAction Stop
            $process.Refresh()
            $processName = $process.ProcessName
            $start = $process.StartTime.ToUniversalTime()
            $filetime = $start.ToFileTimeUtc()
            $path = [string]$process.Path
            if ([string]::IsNullOrEmpty($path) -or [string]::IsNullOrEmpty([string]$cim.ExecutablePath) -or
                (Get-AsdMetricPathKey $path) -cne (Get-AsdMetricPathKey $cim.ExecutablePath)) {
                $readReason = 'cim_image_identity_mismatch'; throw 'Process image changed or was unavailable.'
            }
            if ($null -eq $cim.CreationDate) { $readReason = 'cim_creation_missing'; throw 'CIM creation time was unavailable.' }
            $cimStart = ([DateTime]$cim.CreationDate).ToUniversalTime()
            # CIM has six fractional decimal places (microseconds), FILETIME
            # has seven (100ns). Permit only the lost fractional microsecond;
            # never use this lossy value for registry/previous identity checks.
            if ([Math]::Abs($cimStart.ToFileTimeUtc() - $filetime) -ge 10) {
                $readReason = 'cim_creation_identity_mismatch'; throw 'Process creation changed since the CIM snapshot.'
            }
            $observation = [ordered]@{
                process_id = [int]$candidateId; parent_process_id = [int]$cim.ParentProcessId
                creation_filetime = $filetime.ToString([Globalization.CultureInfo]::InvariantCulture)
                start_time_utc = $start.ToString('o'); cim_creation_utc = $cimStart.ToString('o')
                executable_path = $path; cpu_total_ms = [double]$process.TotalProcessorTime.TotalMilliseconds
                private_bytes = [long]$process.PrivateMemorySize64; working_set_bytes = [long]$process.WorkingSet64
                handle_count = [int]$process.HandleCount; thread_count = [int]$process.Threads.Count
                observed_elapsed_ms = $ElapsedMilliseconds + $clock.Elapsed.TotalMilliseconds
                observed_utc = [DateTime]::UtcNow.ToString('o')
            }
            # A new Process object checks the current PID after property reads.
            # Holding the first object does not make the CIM/metrics atomic.
            $check = Get-Process -Id $candidateId -ErrorAction Stop
            if ($process.HasExited -or $check.StartTime.ToFileTimeUtc() -ne $filetime -or
                (Get-AsdMetricPathKey ([string]$check.Path)) -cne (Get-AsdMetricPathKey $path)) {
                $readReason = 'identity_changed_during_read'; throw 'Process identity changed during metric collection.'
            }
            $verified.Add([pscustomobject]$observation)
        } catch {
            $gaps.Add([pscustomobject]@{
                reason = $readReason; process_id = [int]$candidateId; parent_process_id = [int]$cim.ParentProcessId
                owner_role = $(if (-not [string]::IsNullOrEmpty([string]$cim.ExecutablePath)) {
                    $imageKeys[(Get-AsdMetricPathKey $cim.ExecutablePath)]
                } else { $null })
                # Only already selected candidates are diagnosed. Never export
                # the all-system CIM table, command lines or environments.
                cim_name = [string]$cim.Name; process_name = $processName
                cim_path_present = -not [string]::IsNullOrEmpty([string]$cim.ExecutablePath)
                process_path_present = -not [string]::IsNullOrEmpty([string]$path)
                cim_executable_path = [string]$cim.ExecutablePath; process_executable_path = $path
                error_type = $_.Exception.GetType().FullName; error_hresult = $_.Exception.HResult
            })
        } finally {
            if ($null -ne $check) { $check.Dispose() }
            if ($null -ne $process) { $process.Dispose() }
        }
    }
    $ownership = Resolve-AsdProcessOwnership -Processes $verified.ToArray() -ImageRoles $normalizedImages -RootIdentities $allRoots -PreviousOwned $PreviousOwned
    foreach ($gap in $ownership.gaps) { $gaps.Add($gap) }
    $verifiedById = @{}
    foreach ($row in $verified) { $verifiedById[[string]$row.process_id] = $row }
    $processes = [Collections.Generic.List[object]]::new()
    foreach ($identity in $ownership.owned_identities) {
        $row = $verifiedById[[string]$identity.process_id]
        foreach ($property in $identity.PSObject.Properties) {
            $row | Add-Member -NotePropertyName $property.Name -NotePropertyValue $property.Value -Force
        }
        $processes.Add($row)
    }
    $groups = [Collections.Generic.List[object]]::new()
    foreach ($group in @($processes | Group-Object -Property role, owner_role, context)) {
        $first = $group.Group[0]
        $unknownGaps = @($gaps | Where-Object { $null -eq $_.owner_role -or $_.owner_role -eq $first.owner_role })
        $aggregate = [ordered]@{
            role = $first.role; owner_role = $first.owner_role; context = $first.context
            process_count = $group.Count; observed_rows_read_complete = $unknownGaps.Count -eq 0
            population_complete = $false
        }
        foreach ($metric in @('cpu_total_ms', 'private_bytes', 'working_set_bytes', 'handle_count', 'thread_count')) {
            $sum = 0.0
            foreach ($row in $group.Group) { $sum += $row.$metric }
            $aggregate[$metric] = $(if ($metric -eq 'cpu_total_ms') { [double]$sum } else { [long]$sum })
        }
        $groups.Add([pscustomobject]$aggregate)
    }
    $clock.Stop()
    return [pscustomobject]@{
        schema_version = 1; sample_index = $SampleIndex; sample_utc = $sampleUtc
        sample_elapsed_ms = $ElapsedMilliseconds; sample_duration_ms = $clock.Elapsed.TotalMilliseconds
        cim_query_duration_ms = $cimClock.Elapsed.TotalMilliseconds; sampler_process_id = $PID
        sampler_is_64_bit_process = [Environment]::Is64BitProcess
        candidate_process_count = $candidateIds.Count; observed_process_count = $processes.Count
        processes = @($processes.ToArray()); groups = @($groups.ToArray()); gaps = @($gaps.ToArray())
        owned_identities = @($ownership.owned_identities)
        cpu_semantics = 'cumulative_user_plus_kernel_ms; sampled_deltas_are_lower_bounds'
        working_set_semantics = 'resident_private_and_shared_pages; cross_process_sum_can_double_count'
        private_bytes_semantics = 'Process.PrivateMemorySize64; nonshareable_allocated_bytes'
        completeness = 'non_atomic_sample; between_sample_processes_and_shared_WMI_provider_cost_are_not_captured'
    }
}
