$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'history-desktop-20261123'
$source=[IO.File]::ReadAllText((Join-Path $base 'resume-history-desktop.ps1'));$prefix=$source.Substring(0,$source.IndexOf('$prior='))
. ([ScriptBlock]::Create($prefix))
$f=NativeFrame $sessions[0];$p=$f.layout
$r=@{failure=([IO.File]::ReadAllText((Join-Path $testRoot 'flow-failure.json'))|ConvertFrom-Json);task_state=[string](Get-ScheduledTask AsdHistoryDesktop20261123).State;backend_alternate_hint=$p.alternate_screen;copy_mode=$p.copy_mode;columns=$p.cols;rows=$p.rows;screen_text=@($p.rows_v2|ForEach-Object{($_.runs.text-join'').TrimEnd()});alt_app_started=(Test-Path (Join-Path $testRoot 'work/alt-started.json'));session_count=@($sessions).Count}
$r|ConvertTo-Json -Depth 7|Set-Content (Join-Path $testRoot 'native-hint-inspection.json') -Encoding utf8;$r|ConvertTo-Json -Depth 7
