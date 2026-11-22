$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$script=[ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'windows-history-resize.ps1')))
$argsBase=@{ProductBinary=(Join-Path $base 'copy-final2-product.exe');ProductSHA256='a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0';PsmuxBinary=(Join-Path $base 'history-backend-candidate/psmux.exe');PsmuxSHA256='ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937';ShellBinary=(Join-Path $base 'pwsh7/pwsh.exe')}
foreach($case in @('before','after','direct','idle')){
 $argsCase=@{};foreach($k in $argsBase.Keys){$argsCase[$k]=$argsBase[$k]};$argsCase.ArtifactRoot=Join-Path $base ('history-matrix-'+$case)
 if($case -eq 'after'){$argsCase.OutputTiming='AfterAttach'}
 if($case -in @('direct','idle')){$argsCase.SkipEarlyCopy=$true;$argsCase.SkipSecondaryResize=$true}
 if($case -eq 'direct'){$argsCase.DirectBackend=$true};if($case -eq 'idle'){$argsCase.IdleWriter=$true}
 & $script @argsCase
}
