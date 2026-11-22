$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/m1-history-ui-run2/desktop-v3'
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

$r=Action 19 @{kind='capture';label='resumed-live-list'}
if($r.state.text -match 'history-fixture' -or $r.state.text -notmatch '0 会话'){throw 'Recovery is not at live list'}
$null=Action 20 @{kind='text';text='h'}
Start-Sleep -Milliseconds 600
$r=Action 21 @{kind='capture';label='history-reopened'}
if($r.state.text -notmatch 'history-fixture-1'){throw 'Reopen failed'}
$null=Action 22 @{kind='text';text='q'}
Start-Sleep -Milliseconds 800
$r=Action 23 @{kind='capture';label='shell-return'}
if($r.state.text -notmatch 'PS '){throw 'Quit did not return shell'}
$null=Action 24 @{kind='finish'}
'History reopen, quit and owned window finish passed.'
