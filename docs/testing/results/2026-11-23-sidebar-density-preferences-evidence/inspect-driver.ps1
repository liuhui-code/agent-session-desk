$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'density-20261123'
$tokens=$null;$parseErrors=$null
[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText((Join-Path $base 'flow.ps1')),[ref]$tokens,[ref]$parseErrors)|Out-Null
$errors=@($parseErrors|ForEach-Object{@{id=$_.ErrorId;message=$_.Message;line=$_.Extent.StartLineNumber}})
$task=Get-ScheduledTask -TaskName AsdDensity20261123
$r=@{utc=[DateTime]::UtcNow.ToString('o');task_state=[string]$task.State;ready=(Test-Path (Join-Path $root 'desktop/ready.json'));parser_errors=$errors}
$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'flow-parser-failure.json') -Encoding utf8
$r|ConvertTo-Json -Depth 5
if($task.State -ne 'Running'){throw 'The desktop driver is not currently running'}
