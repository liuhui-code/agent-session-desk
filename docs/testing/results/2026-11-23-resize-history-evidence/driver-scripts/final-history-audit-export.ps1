$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$names=@('history-repro1','history-repro2','history-repro3','history-direct-control','history-direct-control2','history-idle-control','history-tool-official','history-tool-official-final','history-tool-candidate1','history-tool-candidate2','history-matrix-before','history-matrix-after','history-matrix-direct','history-matrix-idle')
$identities=[Collections.Generic.List[object]]::new();$cases=[Collections.Generic.List[object]]::new()
foreach($name in $names){
 $dir=Join-Path $base $name;$r=[IO.File]::ReadAllText((Join-Path $dir 'result.json'))|ConvertFrom-Json
 $identityPath=Join-Path $dir 'identity.json';if(Test-Path $identityPath){$i=[IO.File]::ReadAllText($identityPath)|ConvertFrom-Json;$p=Get-Process -Id $i.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime -or $p.Path -ne $i.path);$identities.Add(@{case=$name;identity=$i;original_ended=$ended})}
 if($r.namespace -and !$r.direct_backend){
  $binary=if($name -eq 'history-tool-candidate1'){Join-Path $base 'psmux-history-candidate.exe'}elseif($name -eq 'history-tool-candidate2' -or $name -like 'history-matrix-*'){Join-Path $base 'history-backend-candidate/psmux.exe'}else{Join-Path $base 'backend/psmux.exe'}
  $listing=(& (Join-Path $base 'copy-final2-product.exe') --backend $binary --runtime-dir (Join-Path $dir 'runtime') --namespace $r.namespace --state-dir (Join-Path $dir 'state') list --json|Out-String).Trim()
  if($LASTEXITCODE -ne 0 -or $listing -ne '[]'){throw 'Registered namespace not empty'}
  $cases.Add(@{case=$name;namespace=$r.namespace;list_json=$listing})
 }
}
$paths=@((Join-Path $base 'psmux-history-candidate.exe'),(Join-Path $base 'history-backend-candidate/psmux.exe'),(Join-Path $base 'history-backend-tests.exe'),(Join-Path $base 'history-fixture.exe'),(Join-Path $base 'copy-final2-product.exe'))
$alive=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath}})
$audit=@{utc=[DateTime]::UtcNow.ToString('o');result='passed';original_fixtures=$identities.ToArray();listed_namespaces=$cases.ToArray();exact_test_images_alive=$alive;official_psmux_sha256=(Get-FileHash (Join-Path $base 'backend/psmux.exe')).Hash;candidate_sha256=(Get-FileHash (Join-Path $base 'history-backend-candidate/psmux.exe')).Hash;diagnostic_sha256=(Get-FileHash (Join-Path $base 'windows-history-resize.ps1')).Hash;scope='read-only combined audit after matrix and native suite ended; no desktop/settings/clipboard changes'}
if(@($identities|Where-Object{!$_.original_ended}).Count -or $alive.Count){$audit.result='failed'}
$target=Join-Path $base 'history-export-20261123';if(Test-Path $target){throw 'Fresh export required'};New-Item -ItemType Directory $target|Out-Null
$audit|ConvertTo-Json -Depth 12|Set-Content (Join-Path $target 'final-audit.json') -Encoding utf8
foreach($name in $names){$dest=Join-Path $target $name;New-Item -ItemType Directory $dest|Out-Null;foreach($file in @('result.json','stages.json','captures.json','identity.json','startup-inspection.json','cleanup-audit.json','failure-audit.json')){$p=Join-Path (Join-Path $base $name) $file;if(Test-Path $p){Copy-Item $p (Join-Path $dest $file)}}}
$suite=Join-Path $base 'history-candidate-native-suite';$dest=Join-Path $target 'native-suite';New-Item -ItemType Directory $dest|Out-Null
foreach($file in @('receipt.json','native-tests.log')){Copy-Item (Join-Path $suite $file) (Join-Path $dest $file)}
Copy-Item (Join-Path $suite 'fixture-evidence') (Join-Path $dest 'fixture-evidence') -Recurse
Copy-Item (Join-Path $base 'windows-history-resize.ps1') (Join-Path $target 'windows-history-resize.ps1')
Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip=Join-Path $base 'history-export-20261123.zip';if(Test-Path $zip){throw 'Fresh archive required'}
[IO.Compression.ZipFile]::CreateFromDirectory($target,$zip)
@{result=$audit.result;original_fixture_count=$identities.Count;original_fixtures_ended=@($identities|Where-Object{$_.original_ended}).Count;listed_namespaces=$cases.Count;exact_test_images_alive=$alive.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash}|ConvertTo-Json
if($audit.result -ne 'passed'){throw 'Final original identities or test images remain'}
