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
$null=Action 2 @{kind='paste';text="Set-Content -LiteralPath 'tab-a.txt' -Value '中文A 😀' -Encoding utf8"}
$null=Action 3 @{kind='capture';label='a-unsubmitted'}
$null=Action 4 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 700
$r=Action 5 @{kind='capture';label='b-after-shortcut'}
$r.state.text
$null=Action 6 @{kind='paste';text="Set-Content -LiteralPath 'tab-b.txt' -Value '中文B 😀' -Encoding utf8"}
$null=Action 7 @{kind='capture';label='b-unsubmitted'}
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/工作区 with spaces'
@{a_exists=(Test-Path (Join-Path $work 'tab-a.txt'));b_exists=(Test-Path (Join-Path $work 'tab-b.txt'))}|ConvertTo-Json
