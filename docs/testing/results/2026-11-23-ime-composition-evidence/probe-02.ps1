$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/ime-20261123/desktop'
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'}
 $r=[IO.File]::ReadAllText($response)|ConvertFrom-Json
 if(!$r.ok){throw ($r|ConvertTo-Json -Depth 6)}
 return $r
}

$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'ime-20261123';$work=Join-Path $testRoot 'work'
$product=Join-Path $base 'list-mouse-ui.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-ime','--state-dir',(Join-Path $testRoot 'state'))
$owned=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
function Preview($label){$raw=(& $product @g preview $owned.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Preview failed'};$v=$raw|ConvertFrom-Json;if($v.session_id -ne $owned.session_id -or $v.run_id -ne $owned.run_id){throw 'Preview target changed'};[IO.File]::WriteAllText((Join-Path $testRoot ("preview-"+$label+'.json')),$raw,[Text.UTF8Encoding]::new($false));return $v.snapshot.text}
$before=Preview 'preedit'
if($before -match 'ni.?hao|你好'){throw 'Unconfirmed composition leaked into backend screen'}
if(Test-Path (Join-Path $work 'ime-confirm.txt')){throw 'Draft was already submitted'}
$null=Action 8 @{kind='keys';keys=' '}
Start-Sleep -Milliseconds 300
$confirmed=Preview 'confirmed'
if([regex]::Matches($confirmed,'你好').Count -ne 1){throw 'Candidate did not commit exactly once'}
if(Test-Path (Join-Path $work 'ime-confirm.txt')){throw 'Candidate selection submitted command'}
$null=Action 9 @{kind='capture';label='candidate-confirmed-public'}
$null=Action 10 @{kind='scan';text='zhongwen'}
$null=Action 11 @{kind='capture';label='cancel-preedit-public'}
$during=Preview 'cancel-preedit'
if($during -ne $confirmed){throw 'Second preedit leaked into backend screen'}
$null=Action 12 @{kind='keys';keys='{ESC}'}
Start-Sleep -Milliseconds 200
$cancelled=Preview 'cancelled'
if($cancelled -ne $confirmed){throw 'Cancel changed confirmed input'}
$null=Action 13 @{kind='capture';label='candidate-cancelled-public'}
$null=Action 14 @{kind='text';text="'"}
if(Test-Path (Join-Path $work 'ime-confirm.txt')){throw 'Closing quote submitted command'}
$null=Action 15 @{kind='keys';keys='{ENTER}'}
Start-Sleep -Milliseconds 400
$actual=[IO.File]::ReadAllText((Join-Path $work 'ime-confirm.txt')).Trim()
if($actual -ne '你好'){throw "Final IME text mismatch: $actual"}
@{result='passed';text=$actual;preedit_reached_backend=$false;candidate_selection_submitted=$false;cancel_preserved_confirmed=$true;explicit_enter=$true}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'terminal-ime-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'terminal-ime-result.json') -Raw -Encoding utf8
