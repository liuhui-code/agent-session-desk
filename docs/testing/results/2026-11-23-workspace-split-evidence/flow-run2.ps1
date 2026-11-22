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
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'layout-20261123'
$null=Action 24 @{kind='text';text='q'}
Start-Sleep -Milliseconds 500
$launch="& '$(Join-Path $base 'layout-ui.exe')' --backend '$(Join-Path $base 'backend/psmux.exe')' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace asd-layout --state-dir '$(Join-Path $testRoot 'state')'"
$null=Action 25 @{kind='paste';text=$launch}
$null=Action 26 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 27 @{kind='keys';keys='{F6}'}
$null=Action 28 @{kind='text';text='<'}
$null=Action 29 @{kind='capture';label='production-key-forty'}
$null=Action 30 @{kind='keys';keys='{F6}'}
$null=Action 31 @{kind='paste';text="Set-Content -LiteralPath 'key-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 32 @{kind='keys';keys='{ENTER}'}
$null=Action 33 @{kind='keys';keys='{F6}'}
function LeftWidth($state){$line=@($state.text -split "`r?`n"|Where-Object{$_.StartsWith('─')})[0];if(!$line){throw 'Sidebar rule missing'};return $line.IndexOf('│')-1}
function Drag([string]$label,[int]$from,[int]$to){
 $task=Get-ScheduledTask -TaskName AsdInjectedDrag3_20261123
 if($task.State -eq 'Running'){throw 'Previous owned drag still running'}
 @{label=$label;from=$from;to=$to}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'drag-request.json') -Encoding utf8
 Start-ScheduledTask -TaskName AsdInjectedDrag3_20261123
 $receipt=Join-Path $testRoot "drag-$label.json";$errorFile=Join-Path $testRoot "drag-$label-error.txt";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $receipt) -and !(Test-Path $errorFile) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(Test-Path $errorFile){throw ([IO.File]::ReadAllText($errorFile))}
 if(!(Test-Path $receipt)){throw 'Owned drag did not finish'}
}
Drag 'fifty' 49 60
$r=Action 34 @{kind='capture';label='production-drag-fifty'}
if((LeftWidth $r.state) -ne 60){throw "Actual divider at $(LeftWidth $r.state), expected 60"}
$null=Action 35 @{kind='keys';keys='{F6}'}
$null=Action 36 @{kind='paste';text="Set-Content -LiteralPath 'drag-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 37 @{kind='keys';keys='{ENTER}'}
$null=Action 38 @{kind='keys';keys='{F6}'}
$null=Action 39 @{kind='text';text='='}
$null=Action 40 @{kind='keys';keys='{F6}'}
$null=Action 41 @{kind='paste';text="Set-Content -LiteralPath 'reset-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 42 @{kind='keys';keys='{ENTER}'}
$null=Action 43 @{kind='paste';text='Clear-Host'}
$null=Action 44 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$r=Action 45 @{kind='capture';label='production-reset-public'}
if((LeftWidth $r.state) -ne 21){throw 'Reset did not restore compact sidebar'}
$null=Action 46 @{kind='keys';keys='{F6}'}
Drag 'sixty' 43 72
$r=Action 47 @{kind='capture';label='production-sixty-public'}
if((LeftWidth $r.state) -ne 72){throw 'Second actual drag not reflected in layout'}
$null=Action 48 @{kind='text';text='q'}
Start-Sleep -Milliseconds 500
$null=Action 49 @{kind='finish'}
$work=Join-Path $testRoot 'work';$values=@{}
foreach($f in @('key','drag','reset')){$raw=[IO.File]::ReadAllText((Join-Path $work "$f-size.txt"));$values[$f]=[int]$raw.Trim()}
if($values.key -ne 67 -or $values.drag -ne 55 -or $values.reset -ne 94){throw "Native size mismatch $($values|ConvertTo-Json -Compress)"}
@{result='passed';native_widths=$values;checks=@('production five-percent key','SendInput mouse drag','native viewport sizes','reset restores auto compact')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
