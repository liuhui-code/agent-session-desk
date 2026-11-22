$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$sharedRoot=Join-Path $base 'tab-resources-20261123';$root=Join-Path $sharedRoot 'run3'
$source=[IO.File]::ReadAllText((Join-Path $root 'source-receipt.json'))|ConvertFrom-Json;$repo=$source.repo
$self=Get-Process -Id $PID
$state=@{status='running';pid=$PID;filetime=$self.StartTime.ToUniversalTime().ToFileTimeUtc().ToString();path=$self.Path;source_commit=$source.commit;started_utc=[DateTime]::UtcNow.ToString('o')}
$state|ConvertTo-Json|Set-Content (Join-Path $root 'runner-state.json') -Encoding utf8
$failed=$null
function Invoke-Check([string]$Name,[string]$File,[hashtable]$NamedParameters=@{}){
 & $File @NamedParameters *> (Join-Path $root ($Name+'.log'))
 if(-not $?){throw ($Name+' failed')}
}
try {
 Invoke-Check 'native-ui-helper-tests' (Join-Path $repo 'tools/ci/test-windows-resource-ui.ps1')
 Invoke-Check 'native-process-metrics-tests' (Join-Path $repo 'tools/ci/test-windows-process-metrics.ps1') @{EvidencePath=(Join-Path $root 'process-metrics-results.json')}
 Invoke-Check 'native-journal-tests' (Join-Path $repo 'tools/ci/test-windows-fixture-journal.ps1')
 & (Join-Path $repo 'tools/ci/windows-resources.ps1') -PsmuxBinary (Join-Path $base 'backend/psmux.exe') -ProductBinary (Join-Path $sharedRoot 'asd-resource.exe') -FixtureBinary (Join-Path $sharedRoot 'asd-fixture-resource.exe') -EvidenceDirectory (Join-Path $root 'evidence') *> (Join-Path $root 'resource-run.log')
 if(-not $?){throw 'Resource script failed'}
} catch {
 $failed=$_|Out-String
 $failed|Set-Content (Join-Path $root 'runner-error.txt') -Encoding utf8
} finally {
 $expected=[IO.File]::ReadAllText((Join-Path $root 'expected-sources.json'))|ConvertFrom-Json;$actual=[ordered]@{}
 foreach($entry in $expected.PSObject.Properties){$hash=(Get-FileHash (Join-Path $repo $entry.Name)).Hash.ToLowerInvariant();$actual[$entry.Name]=$hash;if($hash -ne $entry.Value){$failed='Source changed during native resource execution: '+$entry.Name}}
 $actual|ConvertTo-Json -Depth 4|Set-Content (Join-Path $root 'source-after.json') -Encoding utf8
 $state.status=if($null -eq $failed){'passed'}else{'failed'};$state.completed_utc=[DateTime]::UtcNow.ToString('o');$state.error=$failed
 $state|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'runner-state.json') -Encoding utf8
}
if($null -ne $failed){throw $failed}
