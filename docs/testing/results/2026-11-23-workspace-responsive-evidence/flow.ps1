$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/responsive-20261123/desktop'
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

$ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $ready)){throw 'Driver not ready'}
$null=Action 1 @{kind='paste';text="function prompt { 'shell> ' }; Clear-Host"}
$null=Action 2 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 600
$null=Action 3 @{kind='text';text='Write-Output narrow-owned'}
$null=Action 4 @{kind='resize';width=600;height=520}
Start-Sleep -Milliseconds 500
$null=Action 5 @{kind='keys';keys='{F6}'}
$r=Action 6 @{kind='capture';label='stacked-public'}
if($r.state.text -notmatch '› Responsive Shell' -or $r.state.text -notmatch 'narrow-owned'){throw 'Stacked list and draft not both visible'}
$null=Action 7 @{kind='keys';keys='{ENTER}'}
$r=Action 8 @{kind='capture';label='narrow-terminal-public'}
if($r.state.text -match '› Responsive Shell' -or $r.state.text -notmatch 'narrow-owned'){throw 'Terminal did not expand after Enter'}
$null=Action 9 @{kind='resize';width=400;height=520}
$null=Action 10 @{kind='keys';keys='{F6}'}
$r=Action 11 @{kind='capture';label='single-list-public'}
if($r.state.text -notmatch '› Responsive Shell' -or $r.state.text -match 'narrow-owned'){throw 'Very narrow single list did not hide terminal'}
$null=Action 12 @{kind='keys';keys='{ENTER}'}
$r=Action 13 @{kind='capture';label='very-narrow-terminal-public'}
if($r.state.text -match '› Responsive Shell' -or $r.state.text -notmatch 'narrow-owned'){throw 'Draft did not survive single-list mode'}
$null=Action 14 @{kind='keys';keys='{ENTER}'}
$null=Action 15 @{kind='resize';width=1113;height=657}
$null=Action 16 @{kind='paste';text="Set-Content -LiteralPath 'responsive-input.txt' -Value 'same-run-input'"}
$null=Action 17 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 500
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/responsive-20261123/work'
if([IO.File]::ReadAllText((Join-Path $work 'responsive-input.txt')).Trim() -ne 'same-run-input'){throw 'Restored input not observed'}
$null=Action 18 @{kind='keys';keys='{F6}'}
$null=Action 19 @{kind='text';text='q'}
Start-Sleep -Milliseconds 400
$null=Action 20 @{kind='finish'}
@{result='passed';checks=@('stacked management list and native draft','Enter expands terminal','single-list hides terminal','same draft returns','input after resizing')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path (Split-Path $root) 'flow-result.json') -Encoding utf8
Get-Content (Join-Path (Split-Path $root) 'flow-result.json') -Raw -Encoding utf8
