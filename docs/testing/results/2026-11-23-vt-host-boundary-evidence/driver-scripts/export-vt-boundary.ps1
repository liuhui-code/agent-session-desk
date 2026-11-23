$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$export=Join-Path $base 'vt-boundary-export-20261123'
if(Test-Path $export){throw 'Fresh export required'}
$null=New-Item -ItemType Directory $export
$cases=@('candidate-system1','candidate-system2','candidate-paste','candidate-utf8','candidate-noresize','candidate-sdk','candidate-sdk2','candidate-sdk3','official-system','candidate-sdk4','official-sdk','official-sdk-detached','candidate-sdk-detached')|ForEach-Object{'vt-boundary-'+$_}
$cases+=@('vt-final-system','vt-final-sdk','vt-final-sdk2','vt-final-sdk3')
$audit=@()
foreach($case in $cases){
 $root=Join-Path $base $case;if(!(Test-Path $root)){continue}
 $destination=Join-Path $export $case;$null=New-Item -ItemType Directory $destination
 foreach($name in @('result.json','frames.json','captures.json','identity.json','normal-input.json','alt-input.json','fixture.ps1')){
  $path=Join-Path $root $name;if(Test-Path $path){Copy-Item -LiteralPath $path -Destination (Join-Path $destination $name)}
 }
 $identities=@();$identityPath=Join-Path $root 'identity.json'
 if(Test-Path $identityPath){$identities+=,([IO.File]::ReadAllText($identityPath)|ConvertFrom-Json)}
 $resultPath=Join-Path $root 'result.json'
 if(Test-Path $resultPath){$result=[IO.File]::ReadAllText($resultPath)|ConvertFrom-Json;foreach($key in @('server_identity','loaded_sdk_host')){if($result.$key){$identities+=,$result.$key}}}
 # Private registry is only read to audit exact original server identities.
 foreach($file in @(Get-ChildItem (Join-Path $root 'runtime') -Filter '*.pid' -File -ErrorAction SilentlyContinue)){
  $content=[IO.File]::ReadAllText($file.FullName).Trim()
  if($content -match '^([1-9][0-9]*):([0-9]{17,20})$'){
   $p=Get-Process -Id ([int]$Matches[1]) -ErrorAction SilentlyContinue
   if($p -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq [long]$Matches[2]){throw ('Owned registry server remains in '+$case)}
  }else{throw 'Unrecognized private server identity'}
 }
 $checked=@();foreach($id in $identities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw ('Original process remains in '+$case)};$checked+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended}}
 $audit+=,@{case=$case;result_present=(Test-Path $resultPath);identities=$checked;private_registry_live_originals=0}
}
@{utc=[DateTime]::UtcNow.ToString('o');scope='Only enumerated owned synthetic probes; no global image-name cleanup';cases=$audit}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $export 'final-audit.json') -Encoding utf8
Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip=Join-Path $base 'vt-boundary-export-20261123.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{cases=$audit.Count;bytes=(Get-Item $zip).Length;sha256=(Get-FileHash $zip).Hash;archive=$zip}|ConvertTo-Json
