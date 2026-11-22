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
$null=Action 20 @{kind='paste';text="Write-Output 'close reopen draft'"}
$null=Action 21 @{kind='keys';keys='{F6}'}
$null=Action 22 @{kind='text';text='x'}
Start-Sleep -Milliseconds 600
$null=Action 23 @{kind='capture';label='tab-closed-run-kept'}
$null=Action 24 @{kind='keys';keys='{DOWN}{ENTER}'}
Start-Sleep -Milliseconds 1200
$r=Action 25 @{kind='capture';label='tab-reopened-draft'}
if($r.state.text -notmatch 'close reopen draft'){throw 'Close/reopen lost original draft'}
$null=Action 26 @{kind='keys';keys='^c'}
$null=Action 27 @{kind='resize';width=540;height=360}
Start-Sleep -Milliseconds 1200
$null=Action 28 @{kind='capture';label='narrow-terminal'}
$null=Action 29 @{kind='paste';text="[IO.File]::WriteAllText('native-size.txt', [string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 30 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 800
$null=Action 31 @{kind='resize';width=1113;height=657}
Start-Sleep -Milliseconds 1000
$null=Action 32 @{kind='keys';keys='{F6}'}
$null=Action 33 @{kind='text';text='n'}
Start-Sleep -Milliseconds 800
$r=Action 34 @{kind='capture';label='create-terminal-dialog'}
if($r.state.text -notmatch '新建终端'){throw 'Creation dialog missing'}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$work=Join-Path $base 'tabs-ui-20261123/工作区 with spaces'
$null=Action 35 @{kind='paste';text='普通终端 C'}
$null=Action 36 @{kind='keys';keys='{TAB}'}
$null=Action 37 @{kind='paste';text=(Join-Path $base 'pwsh7/pwsh.exe')}
$null=Action 38 @{kind='keys';keys='{TAB}'}
$null=Action 39 @{kind='paste';text=$work}
$null=Action 40 @{kind='keys';keys='{TAB}'}
$null=Action 41 @{kind='paste';text='["-NoLogo","-NoProfile"]'}
$null=Action 42 @{kind='capture';label='create-terminal-fields'}
$null=Action 43 @{kind='keys';keys='^s'}
Start-Sleep -Milliseconds 1800
$r=Action 44 @{kind='capture';label='created-shell-c'}
if($r.state.text -notmatch '3 会话' -or $r.state.text -notmatch '普通终端 C'){throw 'Created terminal not available'}
$size=[IO.File]::ReadAllText((Join-Path $work 'native-size.txt'))
if([int]$size -lt 1 -or [int]$size -gt 70){throw "Native resize not observed: $size"}
@{native_narrow_width=$size;create='passed';close_reopen='passed'}|ConvertTo-Json
