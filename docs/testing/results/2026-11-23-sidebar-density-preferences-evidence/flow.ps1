$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/density-20261123/desktop'
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
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$testRoot=Join-Path $base 'density-20261123';$work=Join-Path $testRoot 'work'
$preference=Join-Path $testRoot 'state/ui.toml'
$script:sequence=0
function InvokeDesktop($value){$script:sequence++;return Action $script:sequence $value}
function Key($keys){$null=InvokeDesktop @{kind='keys';keys=$keys}}
function Text($text){$null=InvokeDesktop @{kind='text';text=$text}}
function Paste($text){$null=InvokeDesktop @{kind='paste';text=$text}}
function Shot($name){$r=InvokeDesktop @{kind='capture';label=$name};return $r.state.text}
function AssertPreference($value){
 $text=[IO.File]::ReadAllText($preference)
 if($text -notmatch ('sidebar_density = "'+$value+'"')){throw "Preference not saved: $value"}
}
$null=InvokeDesktop @{kind='resize';width=1113;height=657}
Paste "function prompt { 'A> ' }; Clear-Host";Key '{ENTER}'
Text 'Set-Content a-density.txt "A:$PID" #';Text 'S'
$text=Shot 'terminal-literal-s-public'
if($text -notmatch '#S' -or $text -match '侧栏密度'){throw 'Terminal S was consumed as a management shortcut'}
Key '{F6}';Text 'S'
$text=Shot 'settings-compact-public'
if($text -notmatch '>Compact' -or $text -notmatch '侧栏密度'){throw 'Management S did not open the compact choice'}
Key '{LEFT}'
$text=Shot 'settings-full-public'
if($text -notmatch '>Full'){throw 'Left did not choose Full'}
AssertPreference 'full'
Paste 'never sent from settings'
Key '{ESC}'
$text=Shot 'full-cards-public'
if($text -notmatch '│ 可连接' -or $text -notmatch '╰'){throw 'Full cards missing their metadata lines'}
if(Test-Path (Join-Path $work 'a-density.txt')){throw 'Settings implicitly submitted the draft'}
$null=InvokeDesktop @{kind='clickListRow';title='Density Shell B';metadata=$true}
Key '{ENTER}'
Paste "function prompt { 'B> ' }; Clear-Host";Key '{ENTER}'
Text 'Set-Content b-density.txt "B:$PID"'
Key '{F6}';Text 'S'
[IO.File]::SetAttributes($preference,[IO.FileAttributes]::ReadOnly)
try{
 Key '{RIGHT}'
 $until=[DateTime]::UtcNow.AddSeconds(10)
 do{$failure=InvokeDesktop @{kind='state'};if($failure.state.text -match '未保存'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 if($failure.state.text -notmatch '未保存'){throw 'Native read-only save refusal was not visible'}
 AssertPreference 'full'
 if(@(Get-ChildItem (Split-Path $preference) -Filter '.ui-*.tmp' -Force).Count -ne 0){throw 'Failed save left a temporary file'}
}finally{[IO.File]::SetAttributes($preference,[IO.FileAttributes]::Normal)}
Key '{RIGHT}'
$text=Shot 'settings-minimal-public'
if($text -notmatch '>Minimal'){throw 'Minimal choice not visible'}
AssertPreference 'minimal'
Key '{ESC}'
$text=Shot 'minimal-cards-public'
if($text -match '╰ 可连接' -or $text -match '│ 可连接'){throw 'Minimal still uses metadata rows'}
Text 'S';Key ' '
$text=Shot 'settings-auto-public'
if($text -notmatch '>Auto'){throw 'Space did not cycle to Auto'}
AssertPreference 'auto'
Key '{ESC}'
$text=Shot 'auto-wide-public'
if($text -notmatch '│ 可连接'){throw 'Auto did not expand in the wide view'}
$null=InvokeDesktop @{kind='resize';width=600;height=520}
$text=Shot 'auto-stacked-public'
if($text -notmatch '╰ 可连接' -or $text -match '│ 可连接'){throw 'Auto did not reduce density for stacked space'}
$null=InvokeDesktop @{kind='resize';width=400;height=520}
$text=Shot 'auto-single-list-public'
if($text -notmatch '│ 可连接'){throw 'Auto did not expand in the single list'}
$null=InvokeDesktop @{kind='clickListRow';title='Density Shell A';metadata=$true}
Key '{ENTER}'
$text=Shot 'returned-a-draft-public'
if($text -notmatch 'a-density.txt' -or $text -notmatch '#S' -or $text -match 'never sent from settings'){throw 'A draft was changed by settings'}
if((Test-Path (Join-Path $work 'a-density.txt')) -or (Test-Path (Join-Path $work 'b-density.txt'))){throw 'Draft submitted before explicit Enter'}
Key '{ENTER}';Key '^{PGDN}'
$text=Shot 'returned-b-draft-public'
if($text -notmatch 'b-density.txt' -or $text -match 'a-density.txt'){throw 'B draft not restored to its original Run'}
Key '{ENTER}'
Paste 'Set-Content b-after-density.txt "B2:$PID"';Key '{ENTER}'
$values=@{}
foreach($name in @('a-density.txt','b-density.txt','b-after-density.txt')){$values[$name]=[IO.File]::ReadAllText((Join-Path $work $name)).Trim()}
if($values.'a-density.txt' -notmatch '^A:([0-9]+)$'){throw 'A output absent'};$aPid=[int]$Matches[1]
if($values.'b-density.txt' -notmatch '^B:([0-9]+)$'){throw 'B output absent'};$bPid=[int]$Matches[1]
if($aPid -eq $bPid -or $values.'b-after-density.txt' -ne ('B2:'+$bPid)){throw 'Input reached the wrong shell'}
$null=InvokeDesktop @{kind='resize';width=1113;height=657}
Key '{F6}';Text 'S';Key ' '
AssertPreference 'full'
Key '{ESC}'
$text=Shot 'full-before-reopen-public'
if($text -notmatch '管理模式'){throw 'Management focus not observed before quit'}
Text 'q'
$until=[DateTime]::UtcNow.AddSeconds(10)
do{$state=InvokeDesktop @{kind='state'};if($state.state.text -notmatch '(?m)^Agent Session Desk'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
if($state.state.text -match '(?m)^Agent Session Desk'){throw 'UI did not exit normally'}
$backend=Join-Path $base 'backend/psmux.exe';$product=Join-Path $base 'density-ui.exe'
$command="& '$product' --backend '$backend' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace asd-density --state-dir '$(Join-Path $testRoot 'state')'"
Paste $command;Key '{ENTER}'
$until=[DateTime]::UtcNow.AddSeconds(10)
do{$state=InvokeDesktop @{kind='state'};if($state.state.text -match '(?m)^Agent Session Desk' -and $state.state.text -match '终端模式'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
if($state.state.text -notmatch '终端模式'){throw 'Reopened UI did not become ready'}
Key '{F6}';Text 'S'
$text=Shot 'settings-after-ui-reopen-public'
if($text -notmatch '>Full'){throw 'Reopened UI did not restore the persisted choice'}
Key '{ESC}'
$text=Shot 'cards-after-ui-reopen-public'
if($text -notmatch '│ 可连接'){throw 'Reopened UI did not render full cards'}
Text 'q'
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-density','--state-dir',(Join-Path $testRoot 'state'))
$before=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
$after=((& $product @g list --json|Out-String)|ConvertFrom-Json)
if($LASTEXITCODE -ne 0 -or @($after).Count -ne 2){throw 'Sessions did not survive UI restart'}
foreach($v in $after){if(@($before|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Session/Run identity changed'}}
$identities=@();foreach($shellPid in @($aPid,$bPid)){
 $p=Get-Process -Id $shellPid
 if($p.Path -ne (Join-Path $base 'pwsh7/pwsh.exe')){throw 'Unexpected shell executable'}
 $identities+=,@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}
}
$null=InvokeDesktop @{kind='finish'}
@{result='passed';outputs=$values;shell_processes=$identities;before=$before;after=$after;preference=[IO.File]::ReadAllText($preference);checks=@('terminal S remains literal','management-only settings','Full Compact Minimal Auto choice and save','native read-only replacement refusal preserves Full and cleans temp','full card metadata hit testing','Auto current-budget resize','no settings paste or submission','A-B-A original drafts and processes','UI close/reopen restores Full and original Session/Run')}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
