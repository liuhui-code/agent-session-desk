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
$product=Join-Path $base 'list-mouse-ui.exe';$backend=Join-Path $base 'backend/psmux.exe'
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-ime','--state-dir',(Join-Path $testRoot 'state'))
$owned=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
function FormName($state){$match=[regex]::Match($state.text,'名称（可空）[^\r\n]*\r?\n(?<field>[^\r\n]*)');if(!$match.Success){throw 'Name field not observed'};return $match.Groups['field'].Value}
$null=Action 33 @{kind='keys';keys=' '}
$r=Action 34 @{kind='capture';label='form-confirmed-public'}
$confirmedName=FormName $r.state
if([regex]::Matches($confirmedName,'你好').Count -ne 1){throw 'Form candidate did not commit once'}
$null=Action 35 @{kind='scan';text='zhongwen'}
$null=Action 36 @{kind='capture';label='form-cancel-preedit-public'}
$null=Action 37 @{kind='keys';keys='{ESC}'}
$r=Action 38 @{kind='capture';label='form-cancelled-public'}
$cancelledName=FormName $r.state
if($r.state.text -notmatch '新建终端' -or $cancelledName -ne $confirmedName){throw 'IME cancel closed form or changed confirmed text'}
$null=Action 39 @{kind='keys';keys='{ESC}'}
$r=Action 40 @{kind='capture';label='form-dismissed'}
if($r.state.text -match '新建终端'){throw 'Explicit second Escape did not close form'}
$rowsRaw=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=@($rowsRaw|ConvertFrom-Json)
if($rows.Count -ne 1 -or $rows[0].session_id -ne $owned.session_id -or $rows[0].run_id -ne $owned.run_id){throw 'IME actions started or changed Run'}
$null=Action 41 @{kind='text';text='q'}
Start-Sleep -Milliseconds 300
$r=Action 42 @{kind='finish'}
if(!$r.capsRestored){throw 'Caps Lock not restored'}
$control=[IO.File]::ReadAllText((Join-Path $work 'ime-control.txt')).Trim()
if($control -ne '你好'){throw 'Direct control output mismatch'}
@{result='passed';direct_control=$control;form_confirmed='你好';ime_cancel_kept_form=$true;second_escape_closed_form=$true;unchanged_run=$owned;caps_restored=$r.capsRestored}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'control-form-ime-result.json') -Encoding utf8
Get-Content (Join-Path $testRoot 'control-form-ime-result.json') -Raw -Encoding utf8
