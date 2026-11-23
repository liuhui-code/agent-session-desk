$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'bundle-desktop-20261123/desktop'
$done=@(Get-ChildItem (Join-Path $root 'done') -Filter '*.json'|Sort-Object Name);$script:seq=[int]$done[-1].BaseName
function Action($value){$script:seq++;$name=('{0:d3}.json' -f $script:seq);$tmp=Join-Path $root ($name+'.tmp');[IO.File]::WriteAllText($tmp,($value|ConvertTo-Json),[Text.UTF8Encoding]::new($false));Move-Item $tmp (Join-Path $root ('commands/'+$name));$file=Join-Path $root ('responses/'+$name);$until=[DateTime]::UtcNow.AddSeconds(12);while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};$r=[IO.File]::ReadAllText($file)|ConvertFrom-Json;if(!$r.ok){throw 'Driver rejected cleanup action'};return $r}
$state=Action @{kind='state'}
if($state.state.text -notmatch '终端不可交互' -or $state.state.text -notmatch '管理模式'){throw 'Expected refused view in management focus; no keys sent'}
$null=Action @{kind='keys';keys='q'}
$until=[DateTime]::UtcNow.AddSeconds(12)
do{$state=Action @{kind='state'};$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq (Join-Path $base 'bundle-product.exe')});if($live.Count -eq 0 -and $state.state.text -notmatch '(?m)^Agent Session Desk'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
if($live.Count){throw 'UI remains; driver not finished'}
$null=Action @{kind='finish'}
@{result='refused-view-closed-without-terminal-input';sequence=$script:seq}|ConvertTo-Json|Set-Content (Join-Path (Split-Path $root) 'refusal-finish.json') -Encoding utf8
