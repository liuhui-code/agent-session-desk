$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/dialog-fidelity-20261123/desktop'
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
$null=Action 1 @{kind='text';text='n'}
Start-Sleep -Milliseconds 700
$r=Action 2 @{kind='capture';label='dialog-surface-fixed'}
if($r.state.text -notmatch '新建终端'){throw 'Creation dialog missing'}
$null=Action 3 @{kind='paste';text='背景核对'}
$null=Action 4 @{kind='keys';keys='{TAB}'}
$null=Action 5 @{kind='capture';label='dialog-surface-next-focus'}
$null=Action 6 @{kind='resize';width=540;height=360}
Start-Sleep -Milliseconds 700
$null=Action 7 @{kind='capture';label='dialog-narrow'}
$null=Action 8 @{kind='keys';keys='{ESC}'}
$null=Action 9 @{kind='text';text='q'}
Start-Sleep -Milliseconds 500
$null=Action 10 @{kind='finish'}
'Native dialog surface flow passed.'
