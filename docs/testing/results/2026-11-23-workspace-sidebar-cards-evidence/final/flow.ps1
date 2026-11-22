$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/sidebar-final-20261123/desktop'
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
$work=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/sidebar-final-20261123/work'
$null=Action 1 @{kind='paste';text="function prompt { 'A> ' }; Clear-Host"}
$null=Action 2 @{kind='keys';keys='{ENTER}'}
$null=Action 3 @{kind='text';text='Set-Content a-click.txt "A:$PID"'}
$null=Action 4 @{kind='keys';keys='^{PGDN}'}
Start-Sleep -Milliseconds 600
$null=Action 5 @{kind='paste';text="function prompt { 'B> ' }; Clear-Host"}
$null=Action 6 @{kind='keys';keys='{ENTER}'}
$null=Action 7 @{kind='text';text='Set-Content b-click.txt "B:$PID"'}
$null=Action 8 @{kind='keys';keys='{F6}'}
$null=Action 9 @{kind='clickListRow';title='Sidebar Shell A';metadata=$true}
$r=Action 10 @{kind='capture';label='single-selected-a-public'}
if($r.state.text -notmatch '▶ Sidebar Shell A' -or $r.state.text -notmatch 'b-click.txt'){throw 'Single click did not select A while keeping B terminal'}
if((Test-Path (Join-Path $work 'a-click.txt')) -or (Test-Path (Join-Path $work 'b-click.txt'))){throw 'Click implicitly submitted a draft'}
$null=Action 11 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 400
$r=Action 12 @{kind='capture';label='entered-a-public'}
if($r.state.text -notmatch 'a-click.txt' -or $r.state.text -match 'b-click.txt'){throw 'Enter did not reopen A draft'}
$null=Action 13 @{kind='keys';keys='{F6}'}
$null=Action 14 @{kind='clickListRow';title='Sidebar Shell B';double=$true}
Start-Sleep -Milliseconds 400
$r=Action 15 @{kind='capture';label='double-opened-b-public'}
if($r.state.text -notmatch 'b-click.txt' -or $r.state.text -match 'a-click.txt' -or $r.state.text -notmatch '终端模式'){throw 'Double click did not open B'}
if((Test-Path (Join-Path $work 'a-click.txt')) -or (Test-Path (Join-Path $work 'b-click.txt'))){throw 'Double click injected submission'}
$null=Action 16 @{kind='keys';keys='{ENTER}'}
$null=Action 17 @{kind='keys';keys='{F6}'}
$null=Action 18 @{kind='resize';width=600;height=520}
$null=Action 19 @{kind='clickListRow';title='Sidebar Shell A';double=$true}
Start-Sleep -Milliseconds 400
$r=Action 20 @{kind='capture';label='stacked-opened-a-public'}
if($r.state.text -notmatch 'a-click.txt' -or $r.state.text -match '▶ Sidebar Shell A'){throw 'Stacked list double click did not expand A'}
$null=Action 21 @{kind='keys';keys='{ENTER}'}
$null=Action 22 @{kind='keys';keys='{F6}'}
$null=Action 23 @{kind='resize';width=400;height=520}
$null=Action 24 @{kind='clickListRow';title='Sidebar Shell B'}
$r=Action 25 @{kind='capture';label='single-list-selected-b-public'}
if($r.state.text -notmatch '▶ Sidebar Shell B' -or $r.state.text -match 'B>'){throw 'Single-list selection did not stay in management'}
$null=Action 26 @{kind='keys';keys='{ENTER}'}
$null=Action 27 @{kind='paste';text='Set-Content b-after-click.txt "B2:$PID"'}
$null=Action 28 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 400
$values=@{}
foreach($f in @('a-click.txt','b-click.txt','b-after-click.txt')){$values[$f]=[IO.File]::ReadAllText((Join-Path $work $f)).Trim()}
if($values.'a-click.txt' -notmatch '^A:([0-9]+)$'){throw 'A process output missing'};$aPid=[int]$Matches[1]
if($values.'b-click.txt' -notmatch '^B:([0-9]+)$'){throw 'B process output missing'};$bPid=[int]$Matches[1]
if($aPid -eq $bPid -or $values.'b-after-click.txt' -ne ('B2:'+$bPid)){throw 'Inputs did not reach distinct original shell processes'}
$identities=@()
foreach($shellPid in @($aPid,$bPid)){$shell=Get-Process -Id $shellPid;if($shell.Path -ne (Join-Path $env:USERPROFILE 'asd-acceptance-20261122/pwsh7/pwsh.exe')){throw 'Unexpected shell executable'};$identities+=,@{pid=$shell.Id;filetime=$shell.StartTime.ToUniversalTime().ToFileTimeUtc();path=$shell.Path}}
$null=Action 29 @{kind='keys';keys='{F6}'}
$null=Action 30 @{kind='text';text='q'}
Start-Sleep -Milliseconds 400
$null=Action 31 @{kind='finish'}
@{result='passed';outputs=$values;shell_processes=$identities;checks=@('metadata-line click selects its own card without switching','Enter reopens original draft','double click opens without child Enter','stacked and single-list mouse selection','independent shell command results')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path (Split-Path $root) 'flow-result.json') -Encoding utf8
Get-Content (Join-Path (Split-Path $root) 'flow-result.json') -Raw -Encoding utf8
