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
$launch=[IO.File]::ReadAllText((Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/launch.txt')).Replace('tabs-ui.exe','tabs-ui-run2.exe')
$null=Action 72 @{kind='paste';text=$launch}
$null=Action 73 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1200
$null=Action 74 @{kind='capture';label='run2-initial'}
$null=Action 75 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 800
$id=76;$outcomes=@();$failed=$false
for($round=0;$round -lt 6;$round++){
 foreach($size in @(@(540,360),@(1113,657))){
  $null=Action $id @{kind='resize';width=$size[0];height=$size[1]};$id++
  Start-Sleep -Milliseconds 500
  $r=Action $id @{kind='capture';label="run2-size-$round-$($size[0])"};$id++
  $outcomes+=,@{round=$round;width=$size[0];text=$r.state.text}
  if($r.state.text -match '终端不可交互'){$failed=$true;break}
 }
 if($failed){break}
 foreach($focus in @(1,2)){
  $null=Action $id @{kind='keys';keys='{F6}'};$id++
  Start-Sleep -Milliseconds 400
  $r=Action $id @{kind='capture';label="run2-focus-$round-$focus"};$id++
  $outcomes+=,@{round=$round;focus=$focus;text=$r.state.text}
  if($r.state.text -match '终端不可交互'){$failed=$true;break}
 }
 if($failed){break}
}
@{next_id=$id;failed=$failed;observations=$outcomes}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'resize-run2.json') -Encoding utf8
@{next_id=$id;failed=$failed;last_text=$r.state.text}|ConvertTo-Json -Depth 5
