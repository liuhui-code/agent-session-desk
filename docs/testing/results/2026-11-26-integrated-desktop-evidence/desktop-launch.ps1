param([string]$Config)
$ErrorActionPreference='Stop'
$c=Get-Content -Raw $Config|ConvertFrom-Json
if((Test-Path $c.ready)-or (Test-Path $c.exited)){throw 'Fresh launch receipts required'}
@{pid=$PID;filetime=(Get-Process -Id $PID).StartTime.ToUniversalTime().ToFileTimeUtc().ToString()}|ConvertTo-Json|Set-Content $c.ready -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(60)
while(!(Test-Path $c.signal)) {if([DateTime]::UtcNow -gt $until){throw 'No start signal; no manager launched'};Start-Sleep -Milliseconds 100}
& $c.product @($c.globals) '--inside-terminal' '--pet'
$code=$LASTEXITCODE
@{exit_code=$code}|ConvertTo-Json|Set-Content $c.exited -Encoding utf8
exit $code
