$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/desktop'
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
$null=Action 8 @{kind='keys';keys='^{PGUP}'}
Start-Sleep -Milliseconds 500
$null=Action 9 @{kind='capture';label='a-restored-draft'}
$null=Action 10 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 500
$null=Action 11 @{kind='capture';label='a-after-enter'}
$null=Action 12 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 500
$null=Action 13 @{kind='capture';label='b-restored-draft'}
$null=Action 14 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 600
$null=Action 15 @{kind='paste';text="Set-Content -LiteralPath 'tab-b2.txt' -Value 'second command' -Encoding utf8"}
$null=Action 16 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 600
$r=Action 17 @{kind='capture';label='b-two-commands'}
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
$results=@{};foreach($f in @('tab-a.txt','tab-b.txt','tab-b2.txt')){if(!(Test-Path (Join-Path $work $f))){throw "Command not observed: $f"};$results[$f]=[IO.File]::ReadAllText((Join-Path $work $f))}
$results|ConvertTo-Json
