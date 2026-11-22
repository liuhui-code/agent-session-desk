$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
Set-Location (Join-Path $env:USERPROFILE 'asd-acceptance-20261122')
function CheckHash($p,$h) {if((Get-FileHash -LiteralPath $p -Algorithm SHA256).Hash.ToLowerInvariant() -ne $h){throw "Hash mismatch: $p"}}
CheckHash 'asd-windows-amd64-m0-preview.zip' 'ca89f1923a5a1c38a46863c601121fbf45446e917e2e3c9231001d9d23628f1c'
CheckHash 'psmux-v3.4.0-windows-x64.zip' 'a66b6be7c13c6fa0dc3660a267b9e34eaa91a4eb7f07fbd55a0468018bd0f721'
Expand-Archive 'asd-windows-amd64-m0-preview.zip' product
Expand-Archive 'psmux-v3.4.0-windows-x64.zip' backend
$asd=(Resolve-Path 'product/asd-windows-amd64.exe').Path
$psmux=(Get-ChildItem backend -Filter psmux.exe -Recurse | Select-Object -First 1).FullName
CheckHash $asd 'a810705c4322e9f774b61b9159f39340b7681ca4a275d2615b1f8c938fb79cca'
CheckHash $psmux 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'
$common=@('--backend',$psmux,'--runtime-dir',(Join-Path $PWD 'runtime'),'--namespace','asd-ssh-20261122')
function InvokeAsd([string[]]$Options){$out=& $asd @common @Options; if($LASTEXITCODE -ne 0){throw "asd exit $LASTEXITCODE"}; return ($out -join "`n")}
$doctor=InvokeAsd @('doctor') | ConvertFrom-Json
$initial=InvokeAsd @('list','--json') | ConvertFrom-Json
if(@($initial).Count -ne 0){throw 'Refusing to reuse nonempty test namespace'}
$cli='$ErrorActionPreference="Stop"; [Console]::OutputEncoding=[Text.UTF8Encoding]::new(); Write-Output "ASD synthetic interactive CLI ready"; while($null -ne ($line=[Console]::ReadLine())) {if($line -eq "exit"){break}; Write-Output ("ASD received: " + $line)}'
$encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($cli))
$created=@()
foreach($name in @('中文输入 A','会话 B')) {
 $s=InvokeAsd @('start','--name',$name,'--cwd',$PWD.Path,'--',(Join-Path $PSHOME 'powershell.exe'),'-NoLogo','-NoProfile','-EncodedCommand',$encoded) | ConvertFrom-Json
 $created+=,$s
 $created | ConvertTo-Json -Depth 6 | Set-Content created.json -Encoding UTF8
}
$list=InvokeAsd @('list','--json') | ConvertFrom-Json
if(@($list).Count -ne 2 -or $created[0].session_id -eq $created[1].session_id -or $created[0].run_id -eq $created[1].run_id){throw 'Identity/count failure'}
foreach($c in $created){$found=@($list | Where-Object session_id -eq $c.session_id);if($found.Count -ne 1 -or $found[0].run_id -ne $c.run_id -or $found[0].title -cne $c.title){throw 'Metadata roundtrip failure'}}
$record=[ordered]@{utc=[DateTime]::UtcNow.ToString('o');source_commit='5c94b7d407a59be3f8fa476ea6c584aac67110bf';product_sha256=(Get-FileHash $asd).Hash.ToLowerInvariant();psmux_sha256=(Get-FileHash $psmux).Hash.ToLowerInvariant();doctor=$doctor;created=$created;list=$list;case='WIN-UI-02';result='pass';scope='Native Windows CLI via SSH; synthetic Windows PowerShell CLI; no Windows Terminal desktop tested'}
$record | ConvertTo-Json -Depth 8 | Set-Content setup-result.json -Encoding UTF8
Get-Content setup-result.json -Raw
