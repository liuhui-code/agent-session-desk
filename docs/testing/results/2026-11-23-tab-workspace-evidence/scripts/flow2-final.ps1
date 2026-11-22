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
$launch=[IO.File]::ReadAllText((Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/launch.txt')).Replace('tabs-ui.exe','tabs-ui-final.exe')
$null=Action 180 @{kind='paste';text=$launch}
$null=Action 181 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 182 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 1000
$null=Action 183 @{kind='paste';text="Set-Content -LiteralPath 'multiline.txt' -Value @'`nline 1 中文`nline 2`n'@"}
Start-Sleep -Milliseconds 1000
$null=Action 184 @{kind='capture';label='final-multiline-before-enter'}
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
if(Test-Path (Join-Path $work 'multiline.txt')){throw 'Multiline paste submitted before Enter'}
$null=Action 185 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 800
$value=[IO.File]::ReadAllText((Join-Path $work 'multiline.txt'))
if($value.Trim().Replace("`r`n","`n") -ne "line 1 中文`nline 2"){throw 'Multiline contents changed'}
$null=Action 186 @{kind='paste';text='Clear-Host'}
$null=Action 187 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 188 @{kind='capture';label='final-clean-terminal'}
$null=Action 189 @{kind='keys';keys='{F6}'}
$null=Action 190 @{kind='text';text='n'}
Start-Sleep -Milliseconds 600
$null=Action 191 @{kind='capture';label='final-empty-dialog'}
$null=Action 192 @{kind='keys';keys='{ESC}'}
$null=Action 193 @{kind='text';text='?'}
Start-Sleep -Milliseconds 500
$null=Action 194 @{kind='capture';label='final-help'}
$null=Action 195 @{kind='keys';keys='{ESC}'}
Start-Sleep -Milliseconds 500
$null=Action 196 @{kind='text';text='q'}
Start-Sleep -Milliseconds 700
$null=Action 197 @{kind='finish'}
@{result='passed';multiline_value=$value;checks=@('literal multiline without implicit Enter','final dialog','help returns','graceful quit')}|ConvertTo-Json
