$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
foreach($name in @('setup-group-order-desktop.ps1','group-order-desktop-driver.ps1','group-order-flow.ps1')){
 $tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText((Join-Path $base $name),[Text.Encoding]::UTF8),[ref]$tokens,[ref]$errors)
 if($errors.Count){$errors|ForEach-Object{$_.Message};throw 'UTF-8 syntax validation failed'}
}
$r=[IO.File]::ReadAllText((Join-Path $base 'group-order-tests-20261123/outcome.json'))|ConvertFrom-Json
if(@($r.results).Count -ne 3 -or @($r.results|Where-Object{$_.exit_code -ne 0 -or $_.fail -or $_.skip}).Count){throw 'Native test prerequisite failed'}
$tasks=@('AsdGroupOrderSetup20261123','AsdGroupOrderDesktop20261123')
foreach($name in $tasks){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Task already exists'}}
if(Test-Path (Join-Path $base 'group-order-desktop-20261123')){throw 'Desktop root already exists'}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');utf8_syntax_files=3;native_packages=3;owned_tasks_absent=$tasks}|ConvertTo-Json
