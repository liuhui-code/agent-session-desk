$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'integrated-desktop-run2-20261126'
$receipt=Join-Path $out 'result.json'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$product=Join-Path $base 'asd-integrated-r1.exe';$fixture=Join-Path $base 'asd-fixture-integrated.exe'
$artifacts=Get-Content -Raw (Join-Path $base 'integrated-artifacts-r1.json')|ConvertFrom-Json
foreach($item in $artifacts.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw 'Artifact fingerprint mismatch'}}
foreach($entry in @($hostInfo.shell,$hostInfo.terminal)){if((Get-FileHash $entry.path).Hash.ToLowerInvariant() -cne $entry.sha256){throw 'Pinned host changed'}}
if(Test-Path $out){throw 'Fresh isolated directory required'}
New-Item -ItemType Directory $out|Out-Null
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$result=@{result='started';stages=@();clipboard_calls=0;registry_writes=0;personal_state_changed=$false;forced_product_ends=0;mouse_api='SendInput';run='r2'}
$script:originals=@();$script:fixtures=@();$script:windows=@();$script:managers=@();$script:wrappers=@();$script:note=$null;$script:noteWindow=[IntPtr]::Zero;$script:ctrlHeld=$false
$globalArgs=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $out 'runtime'),'--namespace','asd-integrated-r2-20261126','--state-dir',(Join-Path $out 'state'))
$script:window=[IntPtr]::Zero;$script:manager=$null;$script:pet=[IntPtr]::Zero
$cursor=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$cursor)
function Stage([string]$name){$script:stage=$name;$result.stages+=@{stage=$name;utc=[DateTime]::UtcNow.ToString('o')}}
function Wait([string]$label,[scriptblock]$check,[int]$seconds=15){
 $until=[DateTime]::UtcNow.AddSeconds($seconds)
 do{if(& $check){return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw ('Timed out: '+$label+'; no input replay')
}
function Shared([string]$path){
 if(!(Test-Path $path)){return ''}
 $stream=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::ReadWrite-bor[IO.FileShare]::Delete)
 try{if($stream.Length -gt 262144){throw 'Fixture report exceeded bound'};$reader=[IO.StreamReader]::new($stream,[Text.Encoding]::UTF8);try{return $reader.ReadToEnd()}finally{$reader.Dispose()}}finally{$stream.Dispose()}
}
function Journal([string]$path){
 $data=Shared $path;$last=$data.LastIndexOf("`n");if($last -lt 0){return @()}
 return @($data.Substring(0,$last)-split "`r?`n"|Where-Object{$_}|ForEach-Object{$_|ConvertFrom-Json})
}
function Product([string[]]$arguments){
 $text=& $product @globalArgs @arguments 2>&1|Out-String
 if($LASTEXITCODE -ne 0){throw 'Controlled product CLI operation failed'}
 return $text|ConvertFrom-Json
}
function OwnProcess([int]$id,[string]$path){
 $p=Get-Process -Id $id
 $null=$p.Handle
 if($p.Path -ine $path){throw 'Owned process path mismatch'}
 return $p
}
function Surface {
 if($script:window -eq [IntPtr]::Zero){throw 'No owned terminal window'}
 if(![DeskFlow]::Title($script:window).StartsWith($script:title,[StringComparison]::Ordinal)){throw 'Owned terminal title changed'}
 $root=[Windows.Automation.AutomationElement]::FromHandle($script:window)
 $items=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 $best=$null
 foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(65536);if(!$best -or $text.Length -gt $best.text.Length){$best=@{text=$text;pattern=$pattern}}}}
 if($best){return $best};throw 'Owned terminal text pattern unavailable'
}
function Guard {
 if($script:manager -and $script:manager.HasExited){throw 'Owned manager exited'}
 if([DeskFlow]::GetForegroundWindow() -ne $script:window){throw 'Owned terminal lost foreground; no input replay'}
 foreach($key in @(1,2,4,16,17,18,91,92)){if(([DeskFlow]::GetAsyncKeyState($key)-band 32768)-ne 0){throw 'User input held; no injection'}}
 $null=Surface
}
function ClickAt([double]$x,[double]$y){
 $script:lastMouseX=[int]$x;$script:lastMouseY=[int]$y
 if([DeskFlow]::Move([int]$x,[int]$y)-ne 1){throw 'Mouse move incomplete'}
 if([DeskFlow]::Click()-ne 2){throw 'Mouse click incomplete'}
}
function TextPoint([string]$text){
 $v=Surface;$r=$v.pattern.DocumentRange.FindText($text,$false,$false)
 if(!$r){throw 'Controlled text range missing'}
 $rect=@($r.GetBoundingRectangles());if($rect.Count -ne 1){throw 'Controlled text range is clipped or wrapped'}
 return @{x=($rect[0].Left+$rect[0].Width/2);y=($rect[0].Top+$rect[0].Height/2)}
}
function ClickText([string]$text){$point=TextPoint $text;Guard;ClickAt $point.x $point.y}
function Ready([string]$text){Wait 'owned content ready' { $v=Surface;return $v.text.Contains($text)-and !$v.text.Contains('Connecting to')-and !$v.text.Contains('Terminal not interactive')-and !$v.text.Contains('Read-only last') }}
function OpenWorkbench([int]$iteration){
 $script:title='ASD Integrated Flow R2 '+$iteration
 if(@([DeskFlow]::Find($script:title)).Count){throw 'Owned window title already exists'}
 $config=Join-Path $out ('launch-'+$iteration+'.json')
 $ready=Join-Path $out ('host-'+$iteration+'.json');$signal=Join-Path $out ('go-'+$iteration+'.signal');$exited=Join-Path $out ('exit-'+$iteration+'.json')
 @{product=$product;globals=$globalArgs;ready=$ready;signal=$signal;exited=$exited}|ConvertTo-Json -Depth 4|Set-Content $config -Encoding utf8
 $code="& ([ScriptBlock]::Create([IO.File]::ReadAllText('"+(Join-Path $base 'desktop-launch.ps1').Replace("'","''")+"'))) -Config '"+$config.Replace("'","''")+"'"
 $encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
 $info=[Diagnostics.ProcessStartInfo]::new();$info.FileName=$hostInfo.terminal.path;$info.UseShellExecute=$false
 foreach($arg in @('--window','new','new-tab','--title',$script:title,'--suppressApplicationTitle','--startingDirectory',$out,'--',$hostInfo.shell.path,'-NoLogo','-NoProfile','-EncodedCommand',$encoded)){$info.ArgumentList.Add($arg)}
 $launcher=[Diagnostics.Process]::Start($info);$null=$launcher.Handle
 Wait 'private Windows Terminal host ready' {Test-Path $ready}
 $hostReceipt=Get-Content -Raw $ready|ConvertFrom-Json
 $wrapper=OwnProcess $hostReceipt.pid $hostInfo.shell.path;$script:wrappers+=@{process=$wrapper;exited=$exited}
 Wait 'private terminal window' {$matches=@([DeskFlow]::Find($script:title));if($matches.Count -eq 1){$script:window=$matches[0];return $true};return $false}
 $script:windows+=@{handle=$script:window;title=$script:title}
 $null=[DeskFlow]::SetForegroundWindow($script:window);Start-Sleep -Milliseconds 150
 if([DeskFlow]::GetForegroundWindow() -ne $script:window){throw 'Initial private host activation refused; manager not started'}
 $null=Surface
 [IO.File]::WriteAllText($signal,'start')
 Wait 'owned manager and native pet' {
  $matches=@(Get-Process -Name 'asd-integrated-r1' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $product})
  foreach($p in $matches){$pets=@([DeskFlow]::Find('ASD Blocking Reminders',$p.Id));if($pets.Count -eq 1){$script:manager=OwnProcess $p.Id $product;$script:pet=$pets[0];return $true}}
  return $false
 }
 $script:managers+=$script:manager
 Ready 'CLI fixture'
 return $exited
}
function CloseWorkbench([string]$exited){
 ClickText 'Current'
 Wait 'navigation mode' {(Surface).text.Contains('Alt+J/K')}
 Guard;if([DeskFlow]::Text('q')-ne 2){throw 'Quit input incomplete'}
 Wait 'normal manager exit' {$script:manager.HasExited}
 Wait 'wrapper exit receipt' {Test-Path $exited}
 $e=Get-Content -Raw $exited|ConvertFrom-Json
 if($e.exit_code -ne 0 -or $script:manager.ExitCode -ne 0){throw 'Manager did not exit normally'}
 if(@([DeskFlow]::Find('ASD Blocking Reminders',$script:manager.Id)).Count){throw 'Pet survived UI exit'}
 $script:window=[IntPtr]::Zero;$script:manager=$null;$script:pet=[IntPtr]::Zero
}
try{
 Stage 'preflight'
 $desktop=[DeskFlow]::OpenInputDesktop(0,$false,1)
 if($desktop -eq [IntPtr]::Zero -or [DeskFlow]::GetForegroundWindow() -eq [IntPtr]::Zero){throw 'Interactive desktop unavailable'}
 $null=[DeskFlow]::CloseDesktop($desktop)
 $doc=Join-Path $out 'review.md';[IO.File]::WriteAllText($doc,"# ASD_DESKTOP_DOC_v1`r`nControlled Markdown review.`r`n",[Text.UTF8Encoding]::new($false))
 $aLog=Join-Path $out 'current.jsonl';$bLog=Join-Path $out 'decision.jsonl';$form=Join-Path $out 'form.json'
 $a=Product @('start','--name','Current','--cwd',$out,'--env',('ASD_FIXTURE_REPORT_PATH='+$aLog),'--',$fixture);$script:originals+=$a
 Wait 'Current child identity' {@(Journal $aLog).Count -gt 0};$aStart=@(Journal $aLog)[0];$script:fixtures+=OwnProcess $aStart.pid $fixture
 Stage 'initial-workbench'
 $exit1=OpenWorkbench 1
 Wait 'idle pet' {[DeskFlow]::Title($script:pet) -ceq 'ASD Blocking Reminders · idle · 0'}
 ClickText 'CLI fixture'
 Guard;if([DeskFlow]::Text('DRAFT_A_INTEGRATED')-ne 36){throw 'Draft input incomplete'}
 Ready 'DRAFT_A_INTEGRATED'
 Stage 'original-decision-created'
 $b=Product @('start','--name','Decision','--cwd',$out,'--env',('ASD_FIXTURE_REPORT_PATH='+$bLog),'--env',('ASD_FIXTURE_ATTENTION_CONFIG='+$form),'--',$fixture);$script:originals+=$b
 Wait 'Decision child identity' {@(Journal $bLog).Count -gt 0};$bStart=@(Journal $bLog)[0];$script:fixtures+=OwnProcess $bStart.pid $fixture
 $credential=Join-Path $out 'source-credential.json'
 $null=Product @('agent-bind','--session',$b.session_id,'--run',$b.run_id,'--token-file',$credential)
 $attentionID='11111111111111111111111111111111'
 @{CLI=$product;Globals=$globalArgs;CredentialFile=$credential;AttentionID=$attentionID;DocumentFile='review.md'}|ConvertTo-Json -Depth 4|Set-Content $form -Encoding utf8
 Wait 'source pending and pet' {[DeskFlow]::Title($script:pet) -ceq 'ASD Blocking Reminders · approval · 1'}
 if(!(Surface).text.Contains('DRAFT_A_INTEGRATED')){throw 'Passive reminder changed Current view'}
 if(@(Journal $bLog|Where-Object{$_.event -eq 'input'}).Count){throw 'Pending report answered form'}
 Stage 'pet-opens-original-form'
 $rect=[DeskFlow+R]::new();if(![DeskFlow]::GetWindowRect($script:pet,[ref]$rect)){throw 'Pet rectangle missing'}
 Guard;ClickAt (($rect.Left+$rect.Right)/2) ($rect.Top+($rect.Bottom-$rect.Top)*0.92)
 Ready '[y/N]'
 $item=Product @('attention-status',$attentionID)
 if($item.state -cne 'pending' -or !$item.viewed_at -or @(Journal $bLog|Where-Object{$_.event -eq 'input'}).Count){throw 'Viewing resolved or answered form'}
 Stage 'ctrl-click-opens-system-notepad'
 $point=TextPoint 'review.md';Guard
 if([DeskFlow]::Keys([ushort[]]@(0x11),[uint[]]@(0))-ne 1){throw 'Ctrl down incomplete'};$script:ctrlHeld=$true
 try{ClickAt $point.x $point.y}finally{if([DeskFlow]::Keys([ushort[]]@(0x11),[uint[]]@(2))-ne 1){throw 'Ctrl up incomplete'};$script:ctrlHeld=$false}
 Wait 'owned system Notepad process' {
  $matches=@(Get-CimInstance Win32_Process -Property ProcessId,ParentProcessId,ExecutablePath -Filter "Name='notepad.exe'"|Where-Object{$_.ParentProcessId -eq $script:manager.Id -and $_.ExecutablePath -ieq (Join-Path ([Environment]::SystemDirectory) 'notepad.exe')})
  if($matches.Count -eq 1){$script:note=OwnProcess $matches[0].ProcessId $matches[0].ExecutablePath;return $true};return $false
 }
 Wait 'Notepad owned window' {$matches=@([DeskFlow]::Find('',$script:note.Id));if($matches.Count -eq 1){$script:noteWindow=$matches[0];return $true};return $false}
 $root=[Windows.Automation.AutomationElement]::FromHandle($script:noteWindow)
 $nodes=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 $documentSeen=$false
 foreach($node in $nodes){$pattern=$null;if($node.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){if($pattern.DocumentRange.GetText(1024).Contains('ASD_DESKTOP_DOC_v1')){$documentSeen=$true}}}
 if(!$documentSeen){throw 'Notepad did not display controlled Markdown'}
 Wait 'Notepad foreground' {[DeskFlow]::GetForegroundWindow() -eq $script:noteWindow}
 $result.markdown_opened_in_system_notepad=$true
 $result.owner_window_binding_present=([DeskFlow]::OwnerBindings($script:window)-gt 0)
 Stage 'pet-returns-from-notepad'
 if([DeskFlow]::Title($script:pet) -cne 'ASD Blocking Reminders · approval · 1'){throw 'Reading document cleared pending'}
 $rect=[DeskFlow+R]::new();$null=[DeskFlow]::GetWindowRect($script:pet,[ref]$rect)
 foreach($key in @(1,2,4,16,17,18)){if(([DeskFlow]::GetAsyncKeyState($key)-band 32768)-ne 0){throw 'User input held before pet click'}}
 ClickAt (($rect.Left+$rect.Right)/2) ($rect.Top+($rect.Bottom-$rect.Top)*0.92)
 Wait 'pet activates original Windows Terminal' {[DeskFlow]::GetForegroundWindow() -eq $script:window} 8
 Ready '[y/N]';$result.pet_restored_original_terminal_foreground=$true
 $item=Product @('attention-status',$attentionID)
 if($item.state -cne 'pending' -or @(Journal $bLog|Where-Object{$_.event -eq 'input'}).Count){throw 'Pet navigation chose an answer'}
 Stage 'explicit-native-decision'
 Guard;if([DeskFlow]::Text('y')-ne 2){throw 'Explicit y incomplete'}
 Guard;if([DeskFlow]::Key(0x0d)-ne 2){throw 'Explicit Enter incomplete'}
 Wait 'source resolved' {return @(Journal $bLog|Where-Object{$_.event -eq 'decision_resolved'}).Count -eq 1}
 Wait 'idle pet after source resolution' {[DeskFlow]::Title($script:pet) -ceq 'ASD Blocking Reminders · idle · 0'}
 $inputs=@(Journal $bLog|Where-Object{$_.event -eq 'input'})
 if($inputs.Count -ne 1 -or $inputs[0].text -cne 'y'){throw 'Native decision was not exactly one y'}
 $result.decision_only_after_explicit_y_enter=$true
 Stage 'original-draft-and-identities'
 Guard;if([DeskFlow]::Chord(0x12,0x4b)-ne 4){throw 'Alt+K incomplete'}
 Ready 'DRAFT_A_INTEGRATED'
 if(@(Journal $aLog|Where-Object{$_.event -eq 'input'}).Count){throw 'Current draft submitted'}
 $live=@(Product @('list','--json'))
 foreach($original in $script:originals){if(!@($live|Where-Object{$_.session_id -ceq $original.session_id -and $_.run_id -ceq $original.run_id -and $_.backend_id -ceq $original.backend_id}).Count){throw 'Original Session/Run/pane changed'}}
 foreach($p in $script:fixtures){if($p.HasExited){throw 'Original fixture exited before UI quit'}}
 $result.original_runs_and_unsubmitted_draft_preserved=$true
 Stage 'normal-ui-close'
 CloseWorkbench $exit1
 Stage 'reopen-original-workspace'
 $exit2=OpenWorkbench 2
 Ready 'DRAFT_A_INTEGRATED'
 Wait 'reopened idle pet' {[DeskFlow]::Title($script:pet) -ceq 'ASD Blocking Reminders · idle · 0'}
 foreach($p in $script:fixtures){if($p.HasExited){throw 'Original fixture did not survive UI reopen'}}
 if(@(Journal $aLog|Where-Object{$_.event -eq 'input'}).Count){throw 'Reopen submitted Current draft'}
 $result.ui_reopen_preserved_original_fixtures=$true
 CloseWorkbench $exit2
 $result.result='verified'
}catch{try{$surface=Surface;$result.controlled_failure_surface=@{review_reference=$surface.text.Contains('review.md');open_requested=$surface.text.Contains('Markdown open requested');open_error=$surface.text.Contains('Could not open Markdown');owner_binding=([DeskFlow]::OwnerBindings($script:window)-gt 0)};foreach($line in ($surface.text -split "`r?`n")){if($line.Contains('Could not open Markdown')){$result.controlled_open_error=$line.Trim()}}}catch{};$result.result='failed';$result.stage=$script:stage;$result.error=$_.Exception.Message;$result.automatic_input_replay=$false}
finally{
 if($script:ctrlHeld){$null=[DeskFlow]::Keys([ushort[]]@(0x11),[uint[]]@(2))}
 if($script:note -and !$script:note.HasExited -and $script:noteWindow -ne [IntPtr]::Zero){$null=[DeskFlow]::SendMessage($script:noteWindow,0x10,[UIntPtr]::Zero,[IntPtr]::Zero);$result.notepad_closed_normally=$script:note.WaitForExit(10000)}
 foreach($w in $script:windows){if([DeskFlow]::Title($w.handle).StartsWith($w.title,[StringComparison]::Ordinal)){$null=[DeskFlow]::SendMessage($w.handle,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)}}
 foreach($original in $script:originals){try{$null=Product @('stop',$original.session_id)}catch{$result.fixture_cleanup_error=$true}}
 $result.original_fixture_ends=@($script:fixtures|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(10000)};$_.HasExited})
 $result.manager_ends=@($script:managers|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(10000)};$_.HasExited})
 $result.wrapper_ends=@($script:wrappers|ForEach-Object{$p=$_.process;if(!$p.HasExited){$null=$p.WaitForExit(10000)};$p.HasExited})
 $now=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$now)
 if($now.X -eq $script:lastMouseX -and $now.Y -eq $script:lastMouseY){$result.cursor_restored=([DeskFlow]::Move($cursor.X,$cursor.Y)-eq 1)}else{$result.cursor_restored=$false}
 $result.registry_writes=0;$result.product_sha256=(Get-FileHash $product).Hash.ToLowerInvariant()
 $result|ConvertTo-Json -Depth 7|Set-Content $receipt -Encoding utf8
}
