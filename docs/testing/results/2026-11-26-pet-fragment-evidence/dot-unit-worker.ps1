$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'pet-dot-native-20261126';if(Test-Path $out){throw 'Fresh native result required'}
New-Item -ItemType Directory $out|Out-Null
$manifest=Get-Content -Raw (Join-Path $base 'pet-dot-artifacts.json')|ConvertFrom-Json
foreach($file in $manifest.PSObject.Properties){if((Get-FileHash (Join-Path $base $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Artifact fingerprint mismatch'}}
$env:ASD_NATIVE_PET_INTERACTIVE='1'
$pattern='^Test(IdleSpritesHaveNoDetachedInkBelowTheCat|SpriteBandCleanupPreservesBodyPixelsAndCanvas|BlinkKeepsTheCatsHorizontalAnchor|DemoSpritesAndFramesRemainTransparentPremultipliedAndBounded|WindowsIdleBlinkStillPaintsOpenAndClosedFrames|WindowsNativePetGDISurfacesReleaseOriginalObjects)$'
& (Join-Path $base 'pet-dot.test.exe') '-test.v' '-test.run' $pattern '-test.timeout' '30s' 2>&1|Set-Content (Join-Path $out 'pet.stdout') -Encoding utf8
$code=$LASTEXITCODE;$log=@(Get-Content (Join-Path $out 'pet.stdout'))
@{exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count;ui_input_actions=0;clipboard_calls=0;registry_writes=0}|ConvertTo-Json|Set-Content (Join-Path $out 'result.json') -Encoding utf8
