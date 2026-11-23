$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'sidebar-navigation-model-20261123'
if(Test-Path -LiteralPath $root){throw 'Fresh native evidence directory already exists'}
New-Item -ItemType Directory -Path $root|Out-Null
$expected=ConvertFrom-Json '{"sidebar-ui-tests.exe":"4ec352a62c8e999128dbbd8c60a159b94a470bf4b47a7ee534aab8d814102c07","sidebar-config-tests.exe":"72531bec269c37faa2b569021f7134111a222aeefb831cd4ee2566d1dc9b48b0"}'
$results=@()
foreach($name in @('sidebar-ui-tests.exe','sidebar-config-tests.exe')){
 $test=Join-Path $base $name
 $hash=(Get-FileHash -LiteralPath $test).Hash.ToLowerInvariant()
 if($hash -cne $expected.$name){throw 'Transferred test hash differs'}
 $log=Join-Path $root ($name+'.log')
 & $test '-test.v' '-test.count=1' *> $log
 $code=$LASTEXITCODE
 $text=[IO.File]::ReadAllText($log)
 $r=@{name=$name;utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=$hash}
 $results+=$r
 if($code -ne 0 -or $r.fail -or $r.skip){$results|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'results.json') -Encoding utf8;throw 'Native test failed'}
}
$paths=@('sidebar-ui-tests.exe','sidebar-config-tests.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -and $paths -contains $_.ExecutablePath}|Select-Object ProcessId,ExecutablePath,CreationDate)
$result=@{windows=[Environment]::OSVersion.Version.ToString();utc=[DateTime]::UtcNow.ToString('o');tests=$results;live_exact_paths=$live}
$result|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'results.json') -Encoding utf8
$result|ConvertTo-Json -Depth 8
if($live.Count){throw 'Test process remains alive'}
