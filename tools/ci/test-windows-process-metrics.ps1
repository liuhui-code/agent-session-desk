# Reproducible synthetic graph tests; no native process counters are collected.
[CmdletBinding()]
param([string]$EvidencePath = '')

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$helper = Join-Path $PSScriptRoot 'windows-process-metrics.ps1'
. $helper
$fixture = Get-Content -Raw (Join-Path $PSScriptRoot 'process-metrics-fixtures.json') | ConvertFrom-Json
$results = [Collections.Generic.List[object]]::new()
foreach ($case in $fixture.cases) {
    $images = @{}
    foreach ($property in $case.image_roles.PSObject.Properties) { $images[$property.Name] = $property.Value }
    $expectedError = Get-AsdMetricProperty $case 'expect_error'
    try {
        $emitted = @(Resolve-AsdProcessOwnership -Processes $case.processes -ImageRoles $images -RootIdentities $case.roots -PreviousOwned $case.previous)
        if ($null -ne $expectedError) { throw 'An invalid input was unexpectedly accepted.' }
        if ($emitted.Count -ne 1) { throw 'The helper did not emit exactly one snapshot object.' }
        $result = $emitted[0]
        if ($result.owned_identities -isnot [Array] -or $result.gaps -isnot [Array]) { throw 'Array shape was not retained.' }
        if ($result.owned_identities.Count -ne $case.expected.Count) { throw "Wrong number of owned processes: $($result.owned_identities.Count)" }
        foreach ($expected in $case.expected) {
            $actual = @($result.owned_identities | Where-Object { $_.process_id -eq $expected.process_id })
            if ($actual.Count -ne 1) { throw "Expected exactly one matching process $($expected.process_id)" }
            foreach ($field in @('role', 'owner_role', 'is_descendant')) {
                if ($actual[0].$field -cne $expected.$field) { throw "Incorrect $field for process $($expected.process_id): $($actual[0].$field)" }
            }
            if ($actual[0].creation_filetime -isnot [string]) { throw 'FILETIME became a JSON number.' }
        }
        $actualReasons = @($result.gaps | ForEach-Object { $_.reason } | Sort-Object -Unique)
        $expectedReasons = @($case.gap_reasons | Sort-Object -Unique)
        if (($actualReasons -join '|') -cne ($expectedReasons -join '|')) { throw "Unexpected gaps: $($actualReasons -join ',')" }
    } catch {
        if ($null -eq $expectedError -or $_.Exception.Message -cne $expectedError) { throw "Case $($case.name): $($_.Exception.Message)" }
    }
    $results.Add([pscustomobject]@{ name = $case.name; result = 'pass' })
}
foreach ($case in $fixture.path_key_cases) {
    $key = Get-AsdMetricPathKey $case.path
    if ($key -cne $case.expected_key) { throw "Path case $($case.name): unexpected normalized key." }
    $comparisonKey = Get-AsdMetricPathKey $case.compare_to
    if (($key -ceq $comparisonKey) -ne $case.expect_equal) { throw "Path case $($case.name): incorrect path equivalence." }
    $results.Add([pscustomobject]@{ name = $case.name; result = 'pass'; category = 'path_key' })
}
if (-not $IsWindows) {
    $guarded = $false
    try { $null = Get-AsdWindowsProcessSnapshot } catch { $guarded = $_.Exception.Message -eq 'Windows process counters require a native Windows host.' }
    if (-not $guarded) { throw 'Native Windows execution guard failed.' }
    $results.Add([pscustomobject]@{ name = 'linux_native_measurement_guard'; result = 'pass' })
}
$report = [pscustomobject]@{
    schema_version = 1; powershell = $PSVersionTable.PSVersion.ToString(); host_platform = [Runtime.InteropServices.RuntimeInformation]::OSDescription
    helper_sha256 = (Get-FileHash -LiteralPath $helper -Algorithm SHA256).Hash.ToLowerInvariant()
    fixture_sha256 = (Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'process-metrics-fixtures.json') -Algorithm SHA256).Hash.ToLowerInvariant()
    pure_graph_cases = $fixture.cases.Count; path_key_cases = $fixture.path_key_cases.Count
    platform_guard_exercised = -not $IsWindows
    passed = $results.Count; scope = 'Pure synthetic process ownership and lexical Windows path equivalence; non-Windows guard where applicable. No Windows metrics executed.'
    cases = @($results.ToArray())
}
if ($EvidencePath) { $report | ConvertTo-Json -Depth 8 | Set-Content -Encoding utf8 $EvidencePath }
$report | ConvertTo-Json -Depth 8
