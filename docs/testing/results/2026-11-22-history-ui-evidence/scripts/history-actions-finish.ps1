$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/m1-history-ui-run2/desktop-v2'
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

$null=Action 12 @{kind='keys';keys='{HOME}'}
Start-Sleep -Milliseconds 300
$r=Action 13 @{kind='capture';label='history-narrow-top'}
if($r.state.text -notmatch '只读记录' -or $r.state.text -notmatch 'Session'){throw 'Detail Home failed'}
$null=Action 14 @{kind='keys';keys='{ESC}{HOME}{ENTER}'}
$null=Action 15 @{kind='resize';width=950;height=590}
Start-Sleep -Milliseconds 600
$r=Action 16 @{kind='capture';label='history-stop-only'}
if($r.state.text -notmatch 'history-fixture-2' -or $r.state.text -notmatch 'pane 退出观测：未记录' -or $r.state.text -notmatch '停止确认：'){throw 'Stop incorrectly became exit observation'}
$null=Action 17 @{kind='keys';keys='{ESC}{ESC}'}
Start-Sleep -Milliseconds 600
$r=Action 18 @{kind='capture';label='live-return'}
if($r.state.text -match 'history-fixture' -or $r.state.text -notmatch '0 会话'){throw 'History leaked into live list'}
$null=Action 19 @{kind='text';text='h'}
Start-Sleep -Milliseconds 500
$r=Action 20 @{kind='capture';label='history-reopened'}
if($r.state.text -notmatch 'history-fixture-1'){throw 'Reopen failed'}
$null=Action 21 @{kind='text';text='q'}
Start-Sleep -Milliseconds 600
$r=Action 22 @{kind='capture';label='shell-return'}
if($r.state.text -notmatch 'PS '){throw 'Quit did not return shell'}
$null=Action 23 @{kind='finish'}
'Historical stop-only detail, live return, reopen, quit and owned window finish passed.'
