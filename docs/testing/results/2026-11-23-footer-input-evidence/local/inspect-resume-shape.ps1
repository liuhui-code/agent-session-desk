$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/footer-input-desktop-20261123'
$o=[IO.File]::ReadAllText((Join-Path $root 'native-observations.json'))|ConvertFrom-Json
$p=[IO.File]::ReadAllText((Join-Path $root 'preference-observations.json'))|ConvertFrom-Json
$wrapped=@(([IO.File]::ReadAllText((Join-Path $root 'native-observations.json')))|ConvertFrom-Json)
$task=Get-ScheduledTask 'AsdFooterInputDesktop20261123';$driver=Get-Process -Id 10004 -ErrorAction SilentlyContinue
@{utc=[DateTime]::UtcNow.ToString('o');task_state=[string]$task.State;driver_live=[bool]$driver;driver_filetime=$(if($driver){$driver.StartTime.ToUniversalTime().ToFileTimeUtc()});last_response=(Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1).Name;observed_type=$o.GetType().FullName;observed_count=$o.Count;labels=$o.label;wrapped_count=$wrapped.Count;wrapped_first_type=$wrapped[0].GetType().FullName;preferences_count=$p.Count;preference_labels=$p.label}|ConvertTo-Json -Depth 8
