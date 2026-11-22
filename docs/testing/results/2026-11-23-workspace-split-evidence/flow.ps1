$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/layout-20261123/desktop'
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
$null=Action 1 @{kind='paste';text="function prompt { 'shell> ' }; Clear-Host"}
$null=Action 2 @{kind='keys';keys='{ENTER}'}
$null=Action 3 @{kind='text';text='<>'}
Start-Sleep -Milliseconds 1000
$r=Action 4 @{kind='capture';label='terminal-literal-operators'}
if($r.state.text -notmatch 'shell> <>'){throw 'Terminal operators were consumed by manager'}
$null=Action 5 @{kind='keys';keys='{BACKSPACE}{BACKSPACE}{F6}'}
$null=Action 6 @{kind='text';text='<'}
$r=Action 7 @{kind='capture';label='management-forty-percent'}
if($r.state.text -notmatch '会话 40% / 终端 60%'){throw 'Management split key did not adjust five percent'}
$null=Action 8 @{kind='keys';keys='{F6}'}
$null=Action 9 @{kind='paste';text="Set-Content -LiteralPath 'key-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 10 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 500
$null=Action 11 @{kind='keys';keys='{F6}'}
$null=Action 12 @{kind='dragCells';fromx=49;tox=60;row=10}
$r=Action 13 @{kind='capture';label='drag-fifty-percent'}
if($r.state.text -notmatch '会话 50% / 终端 50%'){throw 'Actual mouse drag did not set fifty-percent split'}
$null=Action 14 @{kind='keys';keys='{F6}'}
$null=Action 15 @{kind='paste';text="Set-Content -LiteralPath 'drag-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 16 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 500
$null=Action 17 @{kind='keys';keys='{F6}'}
$null=Action 18 @{kind='text';text='='}
$null=Action 19 @{kind='keys';keys='{F6}'}
$null=Action 20 @{kind='paste';text="Set-Content -LiteralPath 'reset-size.txt' -Value ([string]`$Host.UI.RawUI.WindowSize.Width)"}
$null=Action 21 @{kind='keys';keys='{ENTER}'}
$null=Action 22 @{kind='paste';text='Clear-Host'}
$null=Action 23 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 24 @{kind='capture';label='reset-compact-public'}
$null=Action 25 @{kind='keys';keys='{F6}'}
$null=Action 26 @{kind='dragCells';fromx=43;tox=72;row=10}
$null=Action 27 @{kind='capture';label='manual-sixty-percent-public'}
$null=Action 28 @{kind='text';text='q'}
Start-Sleep -Milliseconds 500
$null=Action 29 @{kind='finish'}
$testRoot=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/layout-20261123';$work=Join-Path $testRoot 'work'
$values=@{};foreach($f in @('key','drag','reset')){$values[$f]=[int]([IO.File]::ReadAllText((Join-Path $work "$f-size.txt"))).Trim()}
if($values.key -ne 67 -or $values.drag -ne 55 -or $values.reset -ne 94){throw "Native size mismatch $($values|ConvertTo-Json -Compress)"}
@{result='passed';native_widths=$values;checks=@('literal terminal operators','management five-percent key','actual mouse drag','native viewport resized','reset restores compact layout')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
