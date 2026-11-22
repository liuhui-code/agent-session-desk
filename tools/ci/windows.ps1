[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string]$PsmuxBinary,
    [string]$EvidenceDirectory = 'build/ci-windows'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
if (-not $IsWindows) { throw 'Native Windows validation must run on Windows.' }

$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$evidence = [IO.Path]::GetFullPath($EvidenceDirectory)
$psmux = (Resolve-Path -LiteralPath $PsmuxBinary).Path
$lock = Get-Content -Raw (Join-Path $PSScriptRoot 'windows-dependencies.json') | ConvertFrom-Json
$actualHash = (Get-FileHash -LiteralPath $psmux -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actualHash -ne $lock.psmux.executable_sha256) {
    throw 'Native acceptance requires the exact pinned official psmux executable.'
}
New-Item -ItemType Directory -Force $evidence | Out-Null
$nativeEvidence = Join-Path $evidence 'native'
New-Item -ItemType Directory -Force $nativeEvidence | Out-Null
$records = [Collections.Generic.List[object]]::new()
$failure = $null
$nativePassed = 0
$sourceBefore = $null
$trackedEnvironment = @('GOFLAGS', 'CGO_ENABLED', 'ASD_PSMUX_TEST_BINARY', 'ASD_PSMUX_TEST_SHA256', 'ASD_FIXTURE_TEST_BINARY', 'ASD_PRODUCT_TEST_BINARY', 'ASD_NATIVE_EVIDENCE_DIR')
$savedEnvironment = @{}
foreach ($name in $trackedEnvironment) {
    $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

function Invoke-Recorded {
    param([string]$Name, [string]$Executable, [string[]]$Arguments)
    $outFile = Join-Path $evidence "$Name.stdout"
    $errFile = Join-Path $evidence "$Name.stderr"
    $clock = [Diagnostics.Stopwatch]::StartNew()
    & $Executable @Arguments 1> $outFile 2> $errFile
    $code = $LASTEXITCODE
    $clock.Stop()
    $records.Add([ordered]@{
        name = $Name; executable = $Executable; arguments = $Arguments
        exit_code = $code; elapsed_ms = $clock.ElapsedMilliseconds
        stdout = "$Name.stdout"; stderr = "$Name.stderr"
    })
    Write-Output "$Name : exit $code ($($clock.ElapsedMilliseconds) ms)"
    if ($code -ne 0) {
        Get-Content -LiteralPath $outFile -Tail 40
        Get-Content -LiteralPath $errFile -Tail 40
        throw "$Name failed with exit code $code"
    }
}

function Get-SourceManifest {
    $hashes = [ordered]@{}
    foreach ($path in (& git ls-files cmd internal go.mod go.sum | Sort-Object)) {
        $hashes[$path] = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant()
    }
    return $hashes
}

Push-Location $repo
try {
    $env:GOFLAGS = '-mod=readonly'
    $env:CGO_ENABLED = '1'
    foreach ($name in $trackedEnvironment | Where-Object { $_ -like 'ASD_*' }) {
        [Environment]::SetEnvironmentVariable($name, $null, 'Process')
    }
    $sourceBefore = Get-SourceManifest
    $sourceBefore | ConvertTo-Json -Depth 5 | Set-Content -Encoding utf8 (Join-Path $evidence 'source-before.json')
    Invoke-Recorded 'psmux-version' $psmux @('-V')
    Invoke-Recorded 'go-version' 'go' @('version')
    Invoke-Recorded 'modules' 'go' @('mod', 'verify')
    Invoke-Recorded 'unit' 'go' @('test', '-count=1', '-timeout=5m', '-json', './...')
    Invoke-Recorded 'race' 'go' @('test', '-race', '-count=1', '-timeout=5m', './...')
    Invoke-Recorded 'vet' 'go' @('vet', './...')
    Invoke-Recorded 'catalog' 'python' @('-B', 'tools/test_catalog.py', 'check', '--generated')
    Invoke-Recorded 'source-size' 'python' @('-B', 'tools/ci/check_source_limits.py')
    $env:CGO_ENABLED = '0'
    $product = Join-Path $repo 'build/asd-windows-native.exe'
    $fixture = Join-Path $repo 'build/asd-fixture-windows-native.exe'
    Invoke-Recorded 'build-product' 'go' @('build', '-trimpath', '-o', $product, './cmd/asd')
    Invoke-Recorded 'build-fixture' 'go' @('build', '-trimpath', '-o', $fixture, './cmd/asd-fixture')
    Invoke-Recorded 'native-config' 'pwsh' @('-NoProfile', '-File', 'tools/ci/windows-config.ps1', '-ProductBinary', $product, '-FixtureBinary', $fixture, '-PsmuxBinary', $psmux, '-EvidenceDirectory', (Join-Path $evidence 'config'))
    $env:ASD_PSMUX_TEST_BINARY = $psmux
    $env:ASD_PSMUX_TEST_SHA256 = $lock.psmux.executable_sha256
    $env:ASD_FIXTURE_TEST_BINARY = $fixture
    $env:ASD_PRODUCT_TEST_BINARY = $product
    $env:ASD_NATIVE_EVIDENCE_DIR = $nativeEvidence
    Invoke-Recorded 'native' 'go' @('test', '-count=1', '-timeout=8m', '-json', '-run', '^TestWindowsPsmuxNative', './internal/backend')
    $events = @(Get-Content -LiteralPath (Join-Path $evidence 'native.stdout') | ForEach-Object { $_ | ConvertFrom-Json })
    $testEvents = @($events | Where-Object { $_.PSObject.Properties.Name -contains 'Test' -and $_.Test -like 'TestWindowsPsmuxNative*' })
    $skipped = @($testEvents | Where-Object { $_.Action -eq 'skip' })
    $passed = @($testEvents | Where-Object { $_.Action -eq 'pass' -and $_.Test -notlike '*/*' })
    $nativePassed = $passed.Count
    if ($skipped.Count -ne 0 -or $nativePassed -lt 5) {
        throw "Native suite did not execute completely: passed=$nativePassed skipped=$($skipped.Count)"
    }
} catch {
    $failure = $_.Exception.Message
    Write-Host "Windows verification failed: $failure"
} finally {
    try {
        # A failed go test still contains useful completed-test events. Keep
        # partial passes visible instead of reporting an artificial zero.
        $nativeOutput = Join-Path $evidence 'native.stdout'
        if (Test-Path -LiteralPath $nativeOutput) {
            try {
                $nativeEvents = @(Get-Content -LiteralPath $nativeOutput | ForEach-Object { $_ | ConvertFrom-Json })
                $nativePassed = @($nativeEvents | Where-Object {
                    $_.PSObject.Properties.Name -contains 'Test' -and
                    $_.Test -like 'TestWindowsPsmuxNative*' -and $_.Test -notlike '*/*' -and $_.Action -eq 'pass'
                }).Count
            } catch {
                if ($null -eq $failure) { $failure = 'Could not parse native test evidence.' }
            }
        }
        $sourceAfter = Get-SourceManifest
        $sourceAfter | ConvertTo-Json -Depth 5 | Set-Content -Encoding utf8 (Join-Path $evidence 'source-after.json')
        if ($null -ne $sourceBefore -and ($sourceBefore | ConvertTo-Json -Compress) -ne ($sourceAfter | ConvertTo-Json -Compress)) {
            $failure = 'Source/module files changed during Windows verification.'
        }
        $os = Get-CimInstance Win32_OperatingSystem
        $summary = [ordered]@{
            schema_version = 1; completed_utc = [DateTime]::UtcNow.ToString('o')
            commit = (& git rev-parse HEAD); runner_os = $env:RUNNER_OS
            image_os = $env:ImageOS; image_version = $env:ImageVersion
            windows_caption = $os.Caption; windows_version = $os.Version; windows_build = $os.BuildNumber
            powershell = $PSVersionTable.PSVersion.ToString(); architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
            psmux_release = $lock.psmux.release_url; psmux_sha256 = $actualHash
            native_tests_passed = $nativePassed; commands = $records.ToArray()
            result = $(if ($null -eq $failure) { 'pass' } else { 'fail' }); failure = $failure
            scope = 'Native Windows process/backend contracts on this runner only; no interactive Windows Terminal, IME, focus, WSL-absence, memory budget or release acceptance.'
        }
        $summary | ConvertTo-Json -Depth 12 | Set-Content -Encoding utf8 (Join-Path $evidence 'summary.json')
    } finally {
        foreach ($name in $trackedEnvironment) {
            [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
        }
        Pop-Location
    }
}
if ($null -ne $failure) { throw $failure }
