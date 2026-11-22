$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$rows=@()
foreach($item in @('backend/psmux.exe','backend-tabs-tests.exe','asd-fixture.exe','pwsh7/pwsh.exe','asd-output.exe')){
 $path=Join-Path $base $item
 $rows+=@{image=$item;sha256=(Get-FileHash $path).Hash.ToLowerInvariant();file_version=(Get-Item $path).VersionInfo.FileVersion}
}
@{utc=[DateTime]::UtcNow.ToString('o');windows_build=[Environment]::OSVersion.Version.ToString();binaries=$rows}|ConvertTo-Json -Depth 5
