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
$launch=[IO.File]::ReadAllText((Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/launch.txt')).Replace('tabs-ui.exe','tabs-ui-run3.exe')
$null=Action 127 @{kind='paste';text=$launch}
$null=Action 128 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 129 @{kind='capture';label='run3-initial'}
$null=Action 130 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 600
$null=Action 131 @{kind='paste';text="1..30 | ForEach-Object { Write-Output ('W'*120) }"}
$null=Action 132 @{kind='keys';keys='{ENTER}'}
$id=133;$observations=@()
for($round=0;$round -lt 5;$round++){
 foreach($size in @(@(540,360),@(1113,657))){
  $null=Action $id @{kind='resize';width=$size[0];height=$size[1]};$id++
  Start-Sleep -Milliseconds 400
  $r=Action $id @{kind='capture';label="run3-size-$round-$($size[0])"};$id++
  if($r.state.text -match '终端不可交互'){throw 'Terminal disconnected during resize'}
  $observations+=,@{round=$round;width=$size[0];text=$r.state.text}
 }
 foreach($focus in @(1,2)){
  $null=Action $id @{kind='keys';keys='{F6}'};$id++
  Start-Sleep -Milliseconds 300
  $r=Action $id @{kind='capture';label="run3-focus-$round-$focus"};$id++
  if($r.state.text -match '终端不可交互'){throw 'Terminal disconnected during focus change'}
  $observations+=,@{round=$round;focus=$focus;text=$r.state.text}
 }
}
$null=Action $id @{kind='paste';text="Set-Content -LiteralPath 'after-resize.txt' -Value 'still interactive'"};$id++
$null=Action $id @{kind='keys';keys='{ENTER}'};$id++
Start-Sleep -Milliseconds 900
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
if([IO.File]::ReadAllText((Join-Path $work 'after-resize.txt')).Trim() -ne 'still interactive'){throw 'Post-resize input not observed'}
@{next_id=$id;result='passed';observations=$observations}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'resize-run3.json') -Encoding utf8
@{next_id=$id;result='passed';observations=$observations.Count}|ConvertTo-Json
