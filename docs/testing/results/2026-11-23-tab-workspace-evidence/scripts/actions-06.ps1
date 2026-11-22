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
$null=Action 24 @{kind='text';text="Write-Output 'draft survives UI restart'"}
$null=Action 25 @{kind='capture';label='draft-before-ui-exit'}
$null=Action 26 @{kind='keys';keys='{F6}'}
$null=Action 27 @{kind='text';text='q'}
Start-Sleep -Milliseconds 600
$null=Action 28 @{kind='capture';label='manager-exited-runs-preserved'}
$null=Action 29 @{kind='finish'}
