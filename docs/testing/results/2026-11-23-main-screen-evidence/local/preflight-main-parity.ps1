$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
foreach($name in @('setup-main-parity-desktop.ps1','main-parity-desktop-driver.ps1','main-parity-flow.ps1')){
 $tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText((Join-Path $base $name),[Text.Encoding]::UTF8),[ref]$tokens,[ref]$errors)
 if($errors.Count){$errors|ForEach-Object{$_.Message};throw 'UTF-8 syntax validation failed'}
}
$r=[IO.File]::ReadAllText((Join-Path $base 'main-parity-tests-final-20261123/outcome.json'))|ConvertFrom-Json
if(@($r.results).Count -ne 1 -or @($r.results|Where-Object{$_.exit_code -ne 0 -or $_.fail -or $_.skip}).Count){throw 'Native test prerequisite failed'}
$tasks=@('AsdMainParitySetup20261123','AsdMainParityDesktop20261123')
foreach($name in $tasks){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Task already exists'}}
if(Test-Path (Join-Path $base 'main-parity-desktop-20261123')){throw 'Desktop root already exists'}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');utf8_syntax_files=3;native_packages=1;owned_tasks_absent=$tasks}|ConvertTo-Json
