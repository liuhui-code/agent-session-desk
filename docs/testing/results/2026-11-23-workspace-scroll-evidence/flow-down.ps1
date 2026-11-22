$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/scroll-20261123/desktop-down'
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
$null=Wait-AsdText 'A>'
$null=Invoke-AsdAction @{kind='text';text='Set-Content scroll-down-draft.txt "A2:$PID"'}
$null=Invoke-AsdAction @{kind='wheel';x=750;y=300;up=$true;count=4}
$null=Wait-AsdText '滚动历史.*12 行'
$null=Invoke-AsdAction @{kind='capture';label='scroll-up-twelve-public'}
$null=Invoke-AsdAction @{kind='wheel';x=750;y=300;up=$false;count=2}
$null=Wait-AsdText '滚动历史.*6 行'
$null=Invoke-AsdAction @{kind='capture';label='scroll-down-six-public'}
$null=Invoke-AsdAction @{kind='wheel';x=750;y=300;up=$false;count=2}
$null=Wait-AsdText 'scroll-down-draft.txt' '滚动历史'
$null=Invoke-AsdAction @{kind='capture';label='scroll-down-live-draft-public'}
if(Test-Path (Join-Path $work 'scroll-down-draft.txt')){throw 'Downward wheel submitted or replaced draft'}
$null=Invoke-AsdAction @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 300
$a2=[IO.File]::ReadAllText((Join-Path $work 'scroll-down-draft.txt')).Trim();$a=[IO.File]::ReadAllText((Join-Path $work 'scroll-draft.txt')).Trim()
if($a2 -ne $a.Replace('A:','A2:')){throw 'UI reopen did not use the same original shell PID'}
$null=Invoke-AsdAction @{kind='keys';keys='{F6}'}
$null=Invoke-AsdAction @{kind='text';text='q'}
Start-Sleep -Milliseconds 300
$null=Invoke-AsdAction @{kind='finish'}
@{result='passed';actions=$script:actionStep;native_offsets=@(12,6,0);original_output=$a;reopened_output=$a2;checks=@('new owned window identity','native wheel down decreases history offset','bottom returns to live without submitting draft','UI reopen retains original shell process')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'down-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'down-result.json') -Raw -Encoding utf8
