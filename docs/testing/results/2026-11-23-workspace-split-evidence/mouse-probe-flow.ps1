$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/layout-20261123/desktop'
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'}
 $r=[IO.File]::ReadAllText($response)|ConvertFrom-Json
 if(!$r.ok){throw ($r|ConvertTo-Json -Depth 6)}
 return $r
}
$null=Action 14 @{kind='text';text='q'}
Start-Sleep -Milliseconds 600
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'layout-20261123'
$code="`$env:ASD_SYNTHETIC_MOUSE_PROBE='$(Join-Path $testRoot 'mouse-probe.log')'; try { & '$(Join-Path $base 'layout-mouse-probe.exe')' --backend '$(Join-Path $base 'backend/psmux.exe')' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace asd-layout --state-dir '$(Join-Path $testRoot 'state')' } finally { Remove-Item env:ASD_SYNTHETIC_MOUSE_PROBE }"
$null=Action 15 @{kind='paste';text=$code}
$null=Action 16 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 17 @{kind='keys';keys='{F6}'}
$null=Action 18 @{kind='text';text='<'}
$null=Action 19 @{kind='capture';label='probe-before-drag'}
$null=Action 20 @{kind='dragCells';fromx=49;tox=60;row=10}
Start-Sleep -Milliseconds 300
$null=Action 21 @{kind='capture';label='probe-after-drag'}
Get-Content (Join-Path $testRoot 'mouse-probe.log') -Raw -Encoding utf8
