$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$r=[IO.File]::ReadAllText((Join-Path $base 'vt-final-sdk2/result.json'))|ConvertFrom-Json
@{server=$r.server_identity;paths=@($r.sdk_host_candidates|ForEach-Object{@{pid=$_.pid;parent=$_.parent_pid;path=[string]$_.path;first_units=@(([string]$_.path).Substring(0,[Math]::Min(8,([string]$_.path).Length)).ToCharArray()|ForEach-Object{[int]$_})}})}|ConvertTo-Json -Depth 7
