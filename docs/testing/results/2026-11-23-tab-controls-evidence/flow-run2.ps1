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
$script:id=7
function Step($value){$r=Action $script:id $value;$script:id++;return $r}
$testRoot=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tab-controls-20261123'
$allRanges=[IO.File]::ReadAllText((Join-Path $testRoot 'text-ranges.json'))|ConvertFrom-Json
$range=@($allRanges|Where-Object{$_.text -match 'Shell A ×'})[0]
if(!$range -or $range.rangeRects.Count -lt 8){throw 'Observed text geometry unavailable'}
$ready=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json
function ClickTab($text,$tail){
 $state=(Step @{kind='state'}).state
 if(($state.rect -join ',') -ne ($ready.rect -join ',')){throw 'Window moved after geometry observation'}
 $row=@($state.text -split "`r?`n"|Where-Object{$_.TrimEnd().EndsWith('+')})[0]
 $column=$row.IndexOf($text);if($column -lt 0){throw "Tab text not visible: $text"}
 if($tail){$column+=$text.Length-1}
 # This bounded run uses ASCII shell titles and a one-cell multiply sign.
 # Width comes from the observed row rectangle, not an assumed pixel font.
 if($row.Substring(0,$column).ToCharArray()|Where-Object{[int]$_ -gt 255}){throw 'Wide-character prefix unsupported by this bounded click harness'}
 $cell=[double]$range.rangeRects[6]/120
 $x=[int]([double]$range.rangeRects[4]+($column+0.5)*$cell-[double]$state.rect[0])
 $y=[int]([double]$range.rangeRects[5]+[double]$range.rangeRects[7]/2-[double]$state.rect[1])
 $null=Step @{kind='click';x=$x;y=$y}
}
$r=Step @{kind='capture';label='before-mouse-fallback'}
$active=[regex]::Match($r.state.text,'›\s+(Shell [AB])').Groups[1].Value
$other='Shell A';if($active -eq $other){$other='Shell B'}
if(!$active -or $r.state.text -notmatch 'draft survives mouse close'){throw 'Original active draft not available'}
ClickTab ($other+' ×') $true
Start-Sleep -Milliseconds 600
$r=Step @{kind='capture';label='inactive-tab-closed'}
if($r.state.text -match ([regex]::Escape($other+' ×')) -or $r.state.text -notmatch 'draft survives mouse close'){throw 'Inactive mouse close changed active draft or kept Tab'}
ClickTab ($active+' ×') $true
Start-Sleep -Milliseconds 600
$r=Step @{kind='capture';label='active-tab-closed'}
if($r.state.text -match 'Shell [AB] ×'){throw 'Last active Tab did not close'}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$product=Join-Path $base 'tab-controls.exe'
$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-tab-controls','--state-dir',(Join-Path $testRoot 'state'))
$rowsText=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$rowsText|ConvertFrom-Json
$expected=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
if(@($rows).Count -ne 2){throw 'Mouse close stopped or duplicated Run'}
foreach($v in $expected){if(@($rows|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Run identity changed'}}
$index=0;for($i=0;$i -lt $rows.Count;$i++){if($rows[$i].title -eq $active){$index=$i;break}}
$null=Step @{kind='keys';keys='{F6}'}
$null=Step @{kind='keys';keys=('{HOME}'+('{DOWN}'*$index)+'{ENTER}')}
Start-Sleep -Milliseconds 1100
$r=Step @{kind='capture';label='original-draft-reopened'}
if($r.state.text -notmatch 'draft survives mouse close'){throw 'Original draft did not survive mouse close/reopen'}
$null=Step @{kind='keys';keys='^c'}
$null=Step @{kind='paste';text='Clear-Host'}
$null=Step @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Step @{kind='capture';label='tab-controls-public'}
ClickTab '+' $false
Start-Sleep -Milliseconds 600
$r=Step @{kind='capture';label='new-tab-dialog-public'}
if($r.state.text -notmatch '新建终端'){throw 'New Tab mouse button did not open form'}
$null=Step @{kind='keys';keys='{ESC}'}
$null=Step @{kind='text';text='q'}
Start-Sleep -Milliseconds 500
$null=Step @{kind='finish'}
@{result='passed';next_id=$script:id;active=$active;other=$other;run_count_after_close=$rows.Count;geometry_method='observed per-row UIA rectangles; ASCII labels and one-cell close mark';checks=@('inactive close keeps active draft','active close keeps Run','list reopens original draft','new Tab opens form without launch')}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
