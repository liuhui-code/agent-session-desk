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
$null=Action 45 @{kind='resize';width=540;height=360}
Start-Sleep -Milliseconds 1200
$null=Action 46 @{kind='paste';text="Set-Content -LiteralPath 'native-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 47 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 800
$null=Action 48 @{kind='capture';label='narrow-size-command'}
$null=Action 49 @{kind='resize';width=1113;height=657}
Start-Sleep -Milliseconds 1200
$null=Action 50 @{kind='paste';text="Set-Content -LiteralPath 'ime.txt' -Value '"}
$null=Action 51 @{kind='physical';text='nihao'}
Start-Sleep -Milliseconds 400
$null=Action 52 @{kind='capture';label='ime-composition-first'}
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
$size=[IO.File]::ReadAllText((Join-Path $work 'native-size.txt'))
if([int]$size -lt 1 -or [int]$size -gt 70){throw "Native resize not observed: $size"}
@{native_narrow_width=$size;result='passed'}|ConvertTo-Json
