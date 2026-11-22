$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/visual-compare-20261123/desktop'
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

$ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $ready)){throw 'Driver not ready'}
$testRoot=Split-Path $root
$last=[int](@(Get-ChildItem (Join-Path $root 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1)[0].BaseName)
$r=Action ($last+1) @{kind='state'}
$r|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'after-connection-failure-state.json') -Encoding utf8
Copy-Item (Join-Path $testRoot 'flow-failure.json') (Join-Path $testRoot 'flow-connection-failure.json')
@{task_state=[string](Get-ScheduledTask -TaskName AsdVisualCompare20261123).State;last_response=$last;state=$r.state.text;failure=([IO.File]::ReadAllText((Join-Path $testRoot 'flow-failure.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 5
