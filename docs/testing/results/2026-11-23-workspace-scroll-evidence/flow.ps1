$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/scroll-20261123/desktop'
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
$script:actionStep=0
function Invoke-AsdAction($request){$script:actionStep++;return Action $script:actionStep $request}
function Wait-AsdText([string]$pattern,[string]$absent=''){
 $limit=[DateTime]::UtcNow.AddSeconds(5)
 do{$r=Invoke-AsdAction @{kind='state'};if($r.state.text -match $pattern -and (!$absent -or $r.state.text -notmatch $absent)){return $r.state};Start-Sleep -Milliseconds 150}while([DateTime]::UtcNow -lt $limit)
 throw "Expected current state $pattern (absent $absent)"
}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'scroll-20261123';$work=Join-Path $testRoot 'work'
$null=Wait-AsdText 'Scroll Shell A.*Scroll Shell B'
$null=Invoke-AsdAction @{kind='paste';text="function prompt { 'A> ' }; Clear-Host"}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
$null=Invoke-AsdAction @{kind='keys';keys='^{PGDN}'}
$null=Wait-AsdText '终端模式'
$null=Invoke-AsdAction @{kind='paste';text="function prompt { 'B> ' }; Clear-Host"}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
$null=Invoke-AsdAction @{kind='keys';keys='^{PGUP}'}
$null=Wait-AsdText 'A>'
$null=Invoke-AsdAction @{kind='paste';text="1..200 | ForEach-Object { 'SCROLL_LINE_{0:D3}' -f `$_ }"}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
$null=Wait-AsdText 'SCROLL_LINE_200'
$null=Invoke-AsdAction @{kind='text';text='Set-Content scroll-draft.txt "A:$PID"'}
$null=Invoke-AsdAction @{kind='capture';label='live-draft-public'}
$null=Invoke-AsdAction @{kind='wheel';x=750;y=300;up=$true;count=4}
$state=Wait-AsdText '滚动历史'
if($state.text -notmatch 'SCROLL_LINE_' -or $state.text -match 'scroll-draft.txt'){throw 'History did not move away from the live draft'}
$null=Invoke-AsdAction @{kind='capture';label='scroll-history-public'}
$null=Invoke-AsdAction @{kind='text';text='/SCROLL_LINE_050'}
$null=Wait-AsdText '滚动历史.*SCROLL_LINE_050'
$null=Invoke-AsdAction @{kind='capture';label='native-search-prompt-public'}
$null=Invoke-AsdAction @{kind='keys';keys='{ESC}'}
$null=Wait-AsdText '滚动历史.*Esc/q'
$null=Invoke-AsdAction @{kind='keys';keys='{ESC}'}
$null=Wait-AsdText 'scroll-draft.txt' '滚动历史'
$null=Invoke-AsdAction @{kind='capture';label='resumed-draft-public'}
if(Test-Path (Join-Path $work 'scroll-draft.txt')){throw 'Scroll/search/Escape submitted the draft'}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 400
$null=Invoke-AsdAction @{kind='wheel';x=750;y=300;up=$true;count=4}
$null=Wait-AsdText '滚动历史'
$null=Invoke-AsdAction @{kind='keys';keys='^{PGDN}'}
$null=Wait-AsdText 'B>' '滚动历史'
$null=Invoke-AsdAction @{kind='paste';text='Set-Content scroll-b.txt "B:$PID"'}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
$null=Invoke-AsdAction @{kind='keys';keys='^{PGUP}'}
$null=Wait-AsdText '滚动历史'
$null=Invoke-AsdAction @{kind='capture';label='reopened-history-public'}
$null=Invoke-AsdAction @{kind='text';text='q'}
$null=Wait-AsdText '终端模式' '滚动历史'
$null=Invoke-AsdAction @{kind='capture';label='q-returned-live-public'}
$product=Join-Path $base 'scroll-ui.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-scroll','--state-dir',(Join-Path $testRoot 'state'))
$owned=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
$a=@($owned|Where-Object{$_.title -eq 'Scroll Shell A'})[0]
$previewRaw=(& $product @g preview $a.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Preview failed'}
$preview=$previewRaw|ConvertFrom-Json;if($preview.run_id -ne $a.run_id){throw 'Run changed'}
if($preview.snapshot.text.TrimEnd() -notmatch 'A>$'){throw 'q leaked into the shell prompt'}
[IO.File]::WriteAllText((Join-Path $testRoot 'q-return-preview.json'),$previewRaw,[Text.UTF8Encoding]::new($false))
$null=Invoke-AsdAction @{kind='keys';keys='{F6}'}
$null=Invoke-AsdAction @{kind='wheel';x=80;y=180;up=$false;count=1}
$null=Wait-AsdText '› Scroll Shell B'
$null=Invoke-AsdAction @{kind='capture';label='list-wheel-selected-b-public'}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
$null=Wait-AsdText 'B>'
$values=@{};foreach($name in @('scroll-draft.txt','scroll-b.txt')){$values[$name]=[IO.File]::ReadAllText((Join-Path $work $name)).Trim()}
if($values.'scroll-draft.txt' -notmatch '^A:([0-9]+)$'){throw 'A output missing'};$aPid=[int]$Matches[1]
if($values.'scroll-b.txt' -notmatch '^B:([0-9]+)$'){throw 'B output missing'};$bPid=[int]$Matches[1]
if($aPid -eq $bPid){throw 'Both commands reached one shell'}
$null=Invoke-AsdAction @{kind='keys';keys='{F6}'}
$null=Invoke-AsdAction @{kind='text';text='q'}
Start-Sleep -Milliseconds 300
$null=Invoke-AsdAction @{kind='finish'}
@{result='passed';actions=$script:actionStep;outputs=$values;checks=@('real wheel enters native scrollback','search prompt visible','Escape preserves and does not submit draft','A-B-A retains per-Run history','q returns to live without child text','list wheel only selects')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
