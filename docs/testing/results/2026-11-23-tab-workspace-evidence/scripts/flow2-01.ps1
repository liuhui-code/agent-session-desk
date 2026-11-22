$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/desktop2'
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
$null=Action 1 @{kind='capture';label='restart-first-tab'}
$null=Action 2 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 1200
$r=Action 3 @{kind='capture';label='restart-draft'}
if($r.state.text -notmatch 'draft survives UI restart'){throw 'UI restart did not preserve Shell draft'}
$null=Action 4 @{kind='keys';keys='^c'}
$null=Action 5 @{kind='keys';keys='^{PGUP}'}
Start-Sleep -Milliseconds 500
$null=Action 6 @{kind='paste';text="Set-Content -LiteralPath 'tab-a.txt' -Value '中文A 😀' -Encoding utf8"}
Start-Sleep -Milliseconds 1200
$r=Action 7 @{kind='capture';label='a-unsubmitted'}
if($r.state.text -notmatch '中文A'){throw 'Paste did not reach terminal'}
$null=Action 8 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 500
$null=Action 9 @{kind='paste';text="Set-Content -LiteralPath 'tab-b.txt' -Value '中文B 😀' -Encoding utf8"}
Start-Sleep -Milliseconds 1200
$null=Action 10 @{kind='capture';label='b-unsubmitted'}
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
if((Test-Path (Join-Path $work 'tab-a.txt')) -or (Test-Path (Join-Path $work 'tab-b.txt'))){throw 'Paste implicitly submitted'}
$null=Action 11 @{kind='keys';keys='^{PGUP}'}
Start-Sleep -Milliseconds 1200
$r=Action 12 @{kind='capture';label='a-restored-draft'}
if($r.state.text -notmatch '中文A' -or $r.state.text -match '中文B'){throw 'A/B draft identity mismatch'}
$null=Action 13 @{kind='keys';keys='{ENTER}'}
$null=Action 14 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 1200
$r=Action 15 @{kind='capture';label='b-restored-draft'}
if($r.state.text -notmatch '中文B' -or $r.state.text -match '中文A'){throw 'B/A draft identity mismatch'}
$null=Action 16 @{kind='keys';keys='{ENTER}'}
$null=Action 17 @{kind='paste';text="Set-Content -LiteralPath 'tab-b2.txt' -Value 'second command' -Encoding utf8"}
$null=Action 18 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 19 @{kind='capture';label='b-two-commands'}
$results=@{};foreach($f in @('tab-a.txt','tab-b.txt','tab-b2.txt')){if(!(Test-Path (Join-Path $work $f))){throw "Command not observed: $f"};$results[$f]=[IO.File]::ReadAllText((Join-Path $work $f))}
$results|ConvertTo-Json
