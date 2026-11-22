$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tab-controls-20261123/desktop'
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
$r=Action 1 @{kind='capture';label='initial-tabs'}
$active=[regex]::Match($r.state.text,'›\s+(Shell [AB])').Groups[1].Value
if(!$active){throw 'Active shell not found'}
$other='Shell A';if($active -eq $other){$other='Shell B'}
$null=Action 2 @{kind='paste';text="function prompt { 'shell> ' }; Clear-Host"}
$null=Action 3 @{kind='keys';keys='{ENTER}'}
$null=Action 4 @{kind='paste';text="Write-Output 'draft survives mouse close'"}
Start-Sleep -Milliseconds 800
$null=Action 5 @{kind='capture';label='draft-before-close'}
$null=Action 6 @{kind='clickText';text=($other+' ×');tail=$true}
Start-Sleep -Milliseconds 700
$r=Action 7 @{kind='capture';label='inactive-tab-closed'}
if($r.state.text -match ([regex]::Escape($other+' ×')) -or $r.state.text -notmatch 'draft survives mouse close'){throw 'Inactive mouse close changed active draft or kept Tab'}
$null=Action 8 @{kind='clickText';text=($active+' ×');tail=$true}
Start-Sleep -Milliseconds 700
$r=Action 9 @{kind='capture';label='active-tab-closed'}
if($r.state.text -match 'Shell [AB] ×'){throw 'Last active Tab did not close'}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$product=Join-Path $base 'tab-controls.exe';$testRoot=Join-Path $base 'tab-controls-20261123'
$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-tab-controls','--state-dir',(Join-Path $testRoot 'state'))
$rowsText=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$rowsText|ConvertFrom-Json
$expected=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
if(@($rows).Count -ne 2){throw 'Mouse close stopped or duplicated Run'}
foreach($v in $expected){if(@($rows|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Run identity changed'}}
$index=0;for($i=0;$i -lt $rows.Count;$i++){if($rows[$i].title -eq $active){$index=$i;break}}
$keys='{HOME}'+('{DOWN}'*$index)+'{ENTER}'
# Both views are closed; the existing focus can still be terminal. F6 selects
# management focus before selecting the original Session from the list.
$null=Action 10 @{kind='keys';keys='{F6}'}
$null=Action 11 @{kind='keys';keys=$keys}
Start-Sleep -Milliseconds 1200
$r=Action 12 @{kind='capture';label='original-draft-reopened'}
if($r.state.text -notmatch 'draft survives mouse close'){throw 'Original draft did not survive mouse close/reopen'}
$null=Action 13 @{kind='keys';keys='^c'}
$null=Action 14 @{kind='paste';text='Clear-Host'}
$null=Action 15 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 16 @{kind='capture';label='tab-controls-public'}
$null=Action 17 @{kind='clickText';text='+'}
Start-Sleep -Milliseconds 700
$r=Action 18 @{kind='capture';label='new-tab-dialog-public'}
if($r.state.text -notmatch '新建终端'){throw 'New Tab mouse button did not open form'}
$null=Action 19 @{kind='keys';keys='{ESC}'}
$null=Action 20 @{kind='text';text='q'}
Start-Sleep -Milliseconds 600
$null=Action 21 @{kind='finish'}
@{result='passed';active=$active;other=$other;run_count_after_close=$rows.Count;checks=@('inactive close keeps active draft','active close keeps Run','list reopens original draft','new Tab opens form without launch')}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
