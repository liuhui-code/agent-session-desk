$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'bundle-desktop-20261123/desktop'
$previous=@(Get-ChildItem (Join-Path $root 'done') -Filter '*.json'|Sort-Object Name);$id=[int]$previous[-1].BaseName+1
$name=('{0:d3}.json' -f $id);$tmp=Join-Path $root ($name+'.tmp');@{kind='resize';width=1750;height=657}|ConvertTo-Json|Set-Content $tmp -Encoding utf8;Move-Item $tmp (Join-Path $root ('commands/'+$name))
$response=Join-Path $root ('responses/'+$name);$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
$id++;$name=('{0:d3}.json' -f $id);@{kind='state'}|ConvertTo-Json|Set-Content (Join-Path $root ($name+'.tmp')) -Encoding utf8;Move-Item (Join-Path $root ($name+'.tmp')) (Join-Path $root ('commands/'+$name))
$response=Join-Path $root ('responses/'+$name);$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
$r=[IO.File]::ReadAllText($response)|ConvertFrom-Json;if(!$r.ok){throw 'Driver refused state inspection'}
$r.state.text
