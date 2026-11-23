$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123'
if((Get-ScheduledTask 'AsdMainParityDesktop20261123').State -eq 'Running'){throw 'Original driver not terminal'}
if(Test-Path (Join-Path $root 'desktop-run2')){throw 'New command directory not fresh'}
foreach($name in @('main-parity-desktop-driver.ps1','main-parity-flow.ps1')){$s=[IO.File]::ReadAllText((Join-Path $base $name),[Text.Encoding]::UTF8);$tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput($s,[ref]$tokens,[ref]$errors);if($errors.Count){throw 'Corrected script syntax error'};if($s -match '\[Windows.Forms.Clipboard\]::'){throw 'Clipboard API still present'}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');clipboard_api_calls=0;new_command_directory='desktop-run2';mutations_replayed=0}|ConvertTo-Json
