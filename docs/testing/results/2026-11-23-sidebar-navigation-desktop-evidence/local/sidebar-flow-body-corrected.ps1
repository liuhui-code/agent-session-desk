$script:preferences=[Collections.Generic.List[object]]::new()
function Pref([string]$label,[string]$density,[string]$mode,[bool]$collapsed){
 $path=Join-Path $testRoot 'state/ui.toml';$until=[DateTime]::UtcNow.AddSeconds(8)
 do{$text=ReadSharedText $path;$ok=$text -match '(?m)^schema_version = 2$' -and $text -match ('(?m)^sidebar_density = "'+$density+'"$') -and $text -match ('(?m)^sidebar_mode = "'+$mode+'"$') -and $text -match ('(?m)^default_group_collapsed = '+$collapsed.ToString().ToLowerInvariant()+'$');if($ok){$script:preferences.Add(@{label=$label;density=$density;mode=$mode;collapsed=$collapsed;text=$text;sha256=(Get-FileHash $path).Hash});$script:preferences.ToArray()|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'preferences.json') -Encoding utf8;return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw ('Preference not observed: '+$label)
}
function ClickNav([string]$title,[bool]$metadata=$false){$null=InvokeVisual @{kind='clickNavRow';title=$title;metadata=$metadata}}
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()})
 if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Pending original draft differs: '+$label)}
}
function ViewCols([int]$cols){
 $text=StateText;$lines=@(($text -split '\r?\n')|Where-Object{$_ -match 'Copy A ×' -and $_ -match 'Copy B ×'});if($lines.Count -ne 1){throw 'Unique application Tab row not observed'};$line=$lines[0]
 if($line.Length -ne $cols){throw ('Actual terminal columns differ: '+$line.Length+' versus '+$cols)}
}
$a="Write-Output ('NAV_A:'+`$PID)";$b="Write-Output ('NAV_B:'+`$PID)"
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Desktop driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 $seed=(Get-FileHash (Join-Path $testRoot 'ui-v1-before.toml')).Hash
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $seed){throw 'UI startup rewrote legacy preferences'}
 Paste $a;Draft 'A-initial-draft' 'Copy A' $a 'NAV_A'
 Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▾ 我的会话 \(2\)';ViewCols 120;$null=Shot 'grouped-compact-120-public'
 Key '1';Key '{ENTER}';$null=WaitText '▸ 我的会话';Pref 'first-collapse' 'compact' 'grouped' $true;$null=Shot 'grouped-folded-120-public'
 Key '%v';$null=WaitText '会话 平铺';Pref 'flat-preserves-fold' 'compact' 'flat' $true
 if((StateText) -match '我的会话'){throw 'Group header remained in flat mode'}
 Key '%v';$null=WaitText '▸ 我的会话';Pref 'return-folded' 'compact' 'grouped' $true
 ClickNav '我的会话';$null=WaitText '▾ 我的会话';Pref 'mouse-expands' 'compact' 'grouped' $false
 ClickNav 'Copy B' $true;$null=WaitText '▶ Copy B';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText 'COPY_READY Copy B'
 Paste $b;Draft 'B-initial-draft' 'Copy B' $b 'NAV_B'
 Key '{F6}';$null=WaitText '管理模式';Key '1';Key '{LEFT}';Key '{LEFT}';$null=WaitText '▸ 我的会话';Pref 'repeat-left' 'compact' 'grouped' $true
 $foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Key '^{PGUP}';$null=WaitText '▾ 我的会话';Draft 'A-tab-reveal' 'Copy A' $a 'NAV_A'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Tab reveal unexpectedly saved expansion'}
 Key '%v';$null=WaitText '会话 平铺';Pref 'explicit-saves-visible-expansion' 'compact' 'flat' $false
 Key '%v';$null=WaitText '▾ 我的会话';Pref 'grouped-expanded' 'compact' 'grouped' $false
 Key 'S';$null=WaitText '侧栏密度';Key '{LEFT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'density-full' 'full' 'grouped' $false;$null=Shot 'grouped-full-120-public'
 ClickNav 'Copy B' $true;$null=WaitText '▶ Copy B';Key '{ENTER}';$null=WaitText '终端模式';Draft 'B-full-metadata-click' 'Copy B' $b 'NAV_B'
 Key '{F6}';$null=WaitText '管理模式';Key 'S';$null=WaitText '侧栏密度';Key '{RIGHT}';Key '{RIGHT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'density-minimal' 'minimal' 'grouped' $false;$null=Shot 'grouped-minimal-120-public'
 Key '1';Key '{DOWN}';$null=WaitText '▶ Copy A';$null=InvokeVisual @{kind='clickNavRow';title='Copy A';wheel=-120};$null=WaitText '▶ Copy B';$null=Shot 'grouped-wheel-selection-public'
 Key 'S';$null=WaitText '侧栏密度';Key '{RIGHT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'density-auto' 'auto' 'grouped' $false
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '▾ 我的会话';ViewCols 63;$null=Shot 'grouped-auto-63-public';ClickNav 'Copy A' $true;$null=WaitText '▶ Copy A'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '▾ 我的会话';ViewCols 49;$null=Shot 'grouped-auto-49-public'
 ClickNav '我的会话';$null=WaitText '▸ 我的会话';Pref 'narrow-mouse-fold' 'auto' 'grouped' $true
 ClickNav '我的会话';$null=WaitText '▾ 我的会话';Pref 'narrow-mouse-expand' 'auto' 'grouped' $false
 ClickNav 'Copy B' $true;$null=WaitText '▶ Copy B';Key '{ENTER}';$null=WaitText '终端模式';Draft 'B-narrow-open' 'Copy B' $b 'NAV_B';$null=Shot 'terminal-49-original-draft-public'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText '终端模式';Key '{F6}';$null=WaitText '管理模式';Key '1';Key '{ENTER}';$null=WaitText '▸ 我的会话';Pref 'before-search-fold' 'auto' 'grouped' $true
 $foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Key '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText '▾ 我的会话';Draft 'A-search-reveal' 'Copy A' $a 'NAV_A'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Search reveal unexpectedly saved expansion'}
 Key '{F6}';$null=WaitText '管理模式';Key 'S';$null=WaitText '侧栏密度';Key '{LEFT}';Key '{LEFT}';Key '{LEFT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'final-density-full' 'full' 'grouped' $false
 Key '1';Key '{ENTER}';Key 'q';WaitExit;Pref 'quit-saved-fold' 'full' 'grouped' $true
 $middle=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $middle.Count -ne 2){throw 'UI quit changed live Session set'}
 foreach($s in $sessions){if(@($middle|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id -and $_.backend_id -eq $s.backend_id}).Count -ne 1){throw 'UI quit changed original Run'}}
 Draft 'A-after-UI-quit' 'Copy A' $a 'NAV_A';Draft 'B-after-UI-quit' 'Copy B' $b 'NAV_B'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▸ 我的会话';Pref 'UI-reopen-retains-fold-density-mode' 'full' 'grouped' $true;$null=Shot 'grouped-full-folded-reopen-public'
 Key '{ENTER}';$null=WaitText '▾ 我的会话';Key '{DOWN}';$null=WaitText '▶ Copy A';Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-reopen-before-enter' 'Copy A' $a 'NAV_A'
 Key '{ENTER}';$null=WaitText ('NAV_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if($oa.capture_text -notmatch ('(?m)^NAV_A:'+$before[0].pid+'\s*$')){throw 'A command ran in another process'}
 Key '^{PGDN}';$null=WaitText '终端模式';Draft 'B-reopen-before-enter' 'Copy B' $b 'NAV_B';Key '{ENTER}';$null=WaitText ('NAV_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if($ob.capture_text -notmatch ('(?m)^NAV_B:'+$before[1].pid+'\s*$')){throw 'B command ran in another process'}
 Key '{F6}';$null=WaitText '管理模式';Key '%v';$null=WaitText '会话 平铺';Pref 'final-flat' 'full' 'flat' $false;$null=Shot 'flat-full-final-public'
 Key 'q';WaitExit;$after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.Count -ne 2){throw 'Final UI quit changed Session set'}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;legacy_read_hash=$seed;before=$sessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();preference_snapshots=$script:preferences.Count;scope='Two ordinary shells; current grouped navigation, 120/63/49 columns, density modes, native mouse/wheel and original drafts across UI reopen. No real agent or full release acceptance.'}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
