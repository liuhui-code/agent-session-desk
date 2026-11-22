$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $b
$asd="$b\product\asd-windows-amd64.exe";$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\desktop-runtime",'--namespace','asd-desktop-20261122')
function Flat($x){if($x -is [Array]){foreach($v in $x){Flat $v}}elseif($x.PSObject.Properties['session_id']){$x}elseif($x.PSObject.Properties['value']){Flat $x.value}else{throw 'Unexpected identity record'}}
$created=@(Flat (Get-Content desktop/created.json -Raw|ConvertFrom-Json))
$live=@(& $asd @common list --json | ConvertFrom-Json)
if($LASTEXITCODE-ne 0){throw 'List failed'}
$live=@(Flat $live)
foreach($c in $created){$x=@($live|Where-Object session_id -eq $c.session_id);if($x.Count-ne 1 -or $x[0].run_id-ne$c.run_id){throw 'Run identity mismatch'}}
$journals=@{}
foreach($n in @('A','B','C','D','E')){$journals[$n]=@(Get-Content "desktop/fixture-$n.jsonl" -Encoding UTF8|ForEach-Object{$_|ConvertFrom-Json})}
$native=@($journals.E|Where-Object kind -eq input)
if($native.Count-ne 2 -or $native[0].line-cne'ASD_NATIVE_1 中文' -or $native[1].line-cne'ASD_NATIVE_2 🧪 é'){throw 'Native Unicode paste mismatch'}
$a=@($journals.A|Where-Object kind -eq input);$bb=@($journals.B|Where-Object kind -eq input)
if($a[0].line-cne'ASD_DESKTOP_A 你好' -or $a[1].line-cne'ASD_AFTER_CANCEL'){throw 'IME commit/cancel mismatch'}
if(@($a|Where-Object {$_.line-match 'ASD_DESKTOP_B|ceshi|测试'}).Count-ne 0){throw 'A cancellation/isolation failure'}
if($bb.Count-ne 1 -or $bb[0].line-cne'ASD_DESKTOP_B'){throw 'B isolation failure'}
$c=$created|Where-Object title -eq '信号测试 C';if(($live|Where-Object session_id -eq $c.session_id).terminal_state-ne'exited'){throw 'CtrlC target not exited'}
$pidEvidence=@(foreach($row in $journals.GetEnumerator()){foreach($start in @($row.Value|Where-Object kind -eq started)){$p=Get-Process -Id $start.pid -ErrorAction SilentlyContinue;@{fixture=$row.Key;pid=$start.pid;alive=[bool]$p;start_time=if($p){$p.StartTime.ToUniversalTime().ToString('o')}else{$null}}}})
@{utc=[DateTime]::UtcNow.ToString('o');created=$created;after_reopen=$live;journals=$journals;fixture_processes=$pidEvidence;identity_unchanged=$true;ime_commit_cancel=$true;input_isolation=$true;native_unicode_paste=$true;ctrl_c_target_exited=$true}|ConvertTo-Json -Depth 12|Set-Content desktop/final-check.json -Encoding UTF8
Get-Content desktop/final-check.json -Raw
