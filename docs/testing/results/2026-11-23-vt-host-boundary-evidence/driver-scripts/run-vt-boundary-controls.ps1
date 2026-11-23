$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$script=[ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'vt-boundary-probe.ps1')))
foreach($case in @('candidate-paste','candidate-utf8','candidate-noresize','candidate-sdk2','official-system')){
 $argsCase=@{Case=$case;ObserveOnly=$true}
 if($case -eq 'candidate-paste'){$argsCase.InputMethod='paste'};if($case -eq 'candidate-utf8'){$argsCase.Utf8Input=$true};if($case -eq 'candidate-noresize'){$argsCase.NoResize=$true};if($case -eq 'candidate-sdk2'){$argsCase.SDK=$true};if($case -eq 'official-system'){$argsCase.Official=$true}
 & $script @argsCase
}
