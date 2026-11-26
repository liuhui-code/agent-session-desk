$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$m=Get-Content -Raw (Join-Path $package 'alpha.json')|ConvertFrom-Json
$target=Join-Path $package $m.entrypoint
$managers=@(Get-Process -Name ([IO.Path]::GetFileNameWithoutExtension($target)) -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
$windows=@([DeskFlow]::Find('ASD Blocking Reminders'))
$demo=@([DeskFlow]::Find('ASD Pet Demo'))
$r=@{manager_count=$managers.Count;blocking_pet_count=$windows.Count;demo_pet_count=$demo.Count;product_sha256=(Get-FileHash $target).Hash.ToLowerInvariant();clipboard_calls=0;ui_actions=0;session_content_exported=0}
if($windows.Count -eq 1){
 $h=$windows[0];$r.pet_idle=([DeskFlow]::Title($h)-ceq 'ASD Blocking Reminders · idle · 0')
 $xs=@();$ys=@()
 for($i=0;$i -lt 80;$i++){$box=[DeskFlow+R]::new();if(![DeskFlow]::GetWindowRect($h,[ref]$box)){throw 'Pet vanished'};$xs+=$box.Left;$ys+=$box.Top;Start-Sleep -Milliseconds 100}
 $r.window_x_range=($xs|Measure-Object -Maximum).Maximum-($xs|Measure-Object -Minimum).Minimum
 $r.window_y_range=($ys|Measure-Object -Maximum).Maximum-($ys|Measure-Object -Minimum).Minimum
}
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$r.bound_host_windows=@([DeskFlow]::HostWindows($hostInfo.terminal.path)|Where-Object{[DeskFlow]::OwnerBindings($_)-gt 0}).Count
$r|ConvertTo-Json|Set-Content (Join-Path $base 'pet-user-probe-20261126.json') -Encoding utf8
