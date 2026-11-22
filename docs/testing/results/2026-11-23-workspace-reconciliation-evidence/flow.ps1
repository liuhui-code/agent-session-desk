$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/workspace-reconcile-20261123/desktop'
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
$r=Action 1 @{kind='capture';label='empty-management'}
if($r.state.text -notmatch '管理模式'){throw 'Empty workspace not in management mode'}
$null=Action 2 @{kind='keys';keys='{F6}'}
$r=Action 3 @{kind='state'}
if($r.state.text -notmatch '管理模式'){throw 'F6 focused nonexistent terminal'}
$null=Action 4 @{kind='text';text='n'}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'workspace-reconcile-20261123'
$null=Action 5 @{kind='paste';text='Reconcile Shell'}
$null=Action 6 @{kind='keys';keys='{TAB}'}
$null=Action 7 @{kind='paste';text=(Join-Path $base 'pwsh7/pwsh.exe')}
$null=Action 8 @{kind='keys';keys='{TAB}'}
$null=Action 9 @{kind='paste';text=(Join-Path $testRoot 'work')}
$null=Action 10 @{kind='keys';keys='{TAB}'}
$null=Action 11 @{kind='paste';text='["-NoLogo","-NoProfile"]'}
$null=Action 12 @{kind='keys';keys='^s'}
Start-Sleep -Milliseconds 1600
$r=Action 13 @{kind='capture';label='created-one-tab'}
$tabRow=@($r.state.text -split "`r?`n"|Where-Object{$_ -match 'Reconcile Shell ×'})[0]
if($r.state.text -notmatch '1 会话' -or ([regex]::Matches($tabRow,'Reconcile Shell')).Count -ne 1){throw 'Create produced wrong session/Tab count'}
$product=Join-Path $base 'reconcile-ui.exe'
$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-ui-reconcile','--state-dir',(Join-Path $testRoot 'state'))
$rowsText=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$rowsText|ConvertFrom-Json
if(@($rows).Count -ne 1){throw 'Wrong created Run count'}
$original=@($rows)[0];$original|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'session.json') -Encoding utf8
$null=Action 14 @{kind='keys';keys='{F6}'}
$null=Action 15 @{kind='text';text='x'}
Start-Sleep -Milliseconds 500
$r=Action 16 @{kind='capture';label='last-tab-closed-management'}
if($r.state.text -notmatch '管理模式' -or $r.state.text -match 'Reconcile Shell ×'){throw 'Last Tab close did not return to management'}
$null=Action 17 @{kind='text';text='n'}
$r=Action 18 @{kind='state'}
if($r.state.text -notmatch '新建终端'){throw 'Empty workspace n not routed to management'}
$null=Action 19 @{kind='keys';keys='{ESC}{HOME}{ENTER}'}
Start-Sleep -Milliseconds 1000
$null=Action 20 @{kind='paste';text="function prompt { 'shell> ' }; Clear-Host"}
$null=Action 21 @{kind='keys';keys='{ENTER}'}
$null=Action 22 @{kind='paste';text="Write-Output 'LAST_OBSERVED_RECONCILE'"}
$null=Action 23 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 1100
$r=Action 24 @{kind='capture';label='last-observed-before-stop'}
if($r.state.text -notmatch '│LAST_OBSERVED_RECONCILE\s*│'){throw 'Actual output marker not observed'}
$beforeText=(& $product @g list --json|Out-String);$before=$beforeText|ConvertFrom-Json
if(@($before|Where-Object{$_.session_id -eq $original.session_id -and $_.run_id -eq $original.run_id}).Count -ne 1){throw 'Reopened Run identity changed'}
$stop=(& $product @g stop $original.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Explicit owned stop failed'}
$stop|Set-Content (Join-Path $testRoot 'stop.json') -Encoding utf8
Start-Sleep -Milliseconds 1200
$r=Action 25 @{kind='capture';label='disconnected-readonly-public'}
if($r.state.text -notmatch '只读' -or $r.state.text -notmatch 'LAST_OBSERVED_RECONCILE'){throw 'Disconnect discarded last observed screen or kept interactive view'}
$null=Action 26 @{kind='keys';keys='{F6}'}
$null=Action 27 @{kind='text';text='q'}
Start-Sleep -Milliseconds 700
$r=Action 28 @{kind='capture';label='readonly-manager-exited'}
if($r.state.text -match 'Agent Session Desk'){throw 'Readonly q did not exit manager'}
$null=Action 29 @{kind='finish'}
@{result='passed';session=$original;checks=@('empty management/F6','one created Session/Tab','last Tab close management','n opens form without active view','same Run reopens','last observed output retained read-only','F6 cannot refocus disconnected terminal','q exits manager')}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
