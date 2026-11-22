$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/desktop'
foreach($f in @('paste-probe-error.txt','paste-probe.json')){if(Test-Path (Join-Path $root $f)){Get-Content (Join-Path $root $f) -Raw -Encoding utf8}}
