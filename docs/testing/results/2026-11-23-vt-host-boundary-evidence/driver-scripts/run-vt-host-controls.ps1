$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$script=[ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'vt-boundary-probe.ps1')))
& $script -Case candidate-sdk3 -SDK -ObserveOnly
& $script -Case official-system -Official -ObserveOnly
