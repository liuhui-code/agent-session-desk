$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/history-desktop-20261123'
$f=[IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json
$observations=[IO.File]::ReadAllText((Join-Path $root 'native-observations.json'))|ConvertFrom-Json
$rows=@();foreach($o in $observations){$rows+=,@{label=$o.label;complete=$o.complete;columns=$o.columns;rows=$o.rows;copy=$o.copy_mode;hint=$o.backend_alternate_hint;controlled_alt=$o.controlled_alt_requested;ids=$o.numbered_ids;screen_text=$o.screen_text}}
$r=@{failure=$f;task_state=[string](Get-ScheduledTask AsdHistoryDesktop20261123).State;observations=$rows}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'progress-inspection.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
