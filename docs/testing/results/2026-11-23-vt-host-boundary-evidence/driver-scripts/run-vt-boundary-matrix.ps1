$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$path=Join-Path $base 'vt-boundary-probe.ps1';$source=[IO.File]::ReadAllText($path)
$tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput($source,[ref]$tokens,[ref]$errors);if($errors.Count){$errors|ForEach-Object{$_.Message};throw 'Probe parser errors'}
$script=[ScriptBlock]::Create($source)
foreach($case in @('candidate-system1','candidate-system2','candidate-sdk','official-system')){
 $argsCase=@{Case=$case;ObserveOnly=$true};if($case -eq 'candidate-sdk'){$argsCase.SDK=$true};if($case -eq 'official-system'){$argsCase.Official=$true}
 & $script @argsCase
}
