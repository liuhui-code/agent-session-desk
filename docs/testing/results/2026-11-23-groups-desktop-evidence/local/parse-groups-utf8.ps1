$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
foreach($name in @('groups-desktop-driver.ps1','groups-flow.ps1')){
 $source=[IO.File]::ReadAllText((Join-Path $base $name),[Text.Encoding]::UTF8);$tokens=$null;$errors=$null
 $null=[Management.Automation.Language.Parser]::ParseInput($source,[ref]$tokens,[ref]$errors)
 @{name=$name;errors=@($errors|ForEach-Object{@{message=$_.Message;line=$_.Extent.StartLineNumber;text=$_.Extent.Text}})}|ConvertTo-Json -Depth 5
}
