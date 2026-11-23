$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$script=[ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'vt-boundary-probe.ps1')))
& $script -Case candidate-sdk4 -SDK -ObserveOnly
& $script -Case official-sdk -Official -SDK -ObserveOnly
& $script -Case official-sdk-detached -Official -SDK -BeforeAttachSeed -ObserveOnly
& $script -Case candidate-sdk-detached -SDK -BeforeAttachSeed -ObserveOnly
