$script:geometry=[Collections.Generic.List[object]]::new()
function Geometry([int]$columns){
 $lines=(StateText) -split '\r?\n';$top=-1
 for($i=0;$i -lt $lines.Length;$i++){if($lines[$i] -match '╭─+╮'){$top=$i;break}}
 if($top -lt 0){throw 'Surface dialog not observed'}
 $line=$lines[$top];$width=$line.IndexOf('╮')-$line.IndexOf('╭')+1;$want=46;if($columns -eq 49){$want=41}
 if($line.Length -ne $columns -or $width -ne $want -or $lines[$top+9] -notmatch '╰─+╯'){throw 'Surface correction changed source geometry'}
 $script:geometry.Add(@{columns=$columns;width=$width;x=$line.IndexOf('╭');top=$top;bottom=$top+9})
}
$draft='Write-Output (''RENAME_STYLE:''+$PID)'
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Surface driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 Paste $draft;CheckDraft 'surface-before-dialog' 'Copy A' $draft 'RENAME_STYLE'
 Key '{F6}';$null=WaitText '管理模式';Text 'r';$null=WaitText '重命名会话';Geometry 120;$null=Shot 'rename-surface-120-public'
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '> Copy A';Geometry 63;$null=Shot 'rename-surface-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '> Copy A';Geometry 49;$null=Shot 'rename-surface-49-public'
 Paste "invalid`nname";$null=WaitText '单行';$null=Shot 'rename-surface-error-49-public'
 Key '{ESC}';$null=WaitText '管理模式';$null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A'
 CheckDraft 'surface-after-cancel' 'Copy A' $draft 'RENAME_STYLE'
 Key '{F6}';$null=WaitText '终端模式';Key '{ENTER}';$null=WaitText ('RENAME_STYLE:'+$before[0].pid)
 $o=Observe 'surface-explicit-enter' 'Copy A';if($o.capture_text -notmatch ('(?m)^RENAME_STYLE:'+$before[0].pid+'\s*$')){throw 'Style dialog changed original draft target'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.session_id -ne $sessions.session_id -or $after.run_id -ne $sessions.run_id -or $after.title -cne 'Copy A'){throw 'Style-only cancel changed Session/Run/name'}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$sessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();geometry=$script:geometry.ToArray();scope='Final style-only binary, one fresh ordinary shell, 120/63/49 cells, invalid paste/cancel and original draft/explicit Enter; full three-shell rename flow belongs to preceding build'}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'style-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'style-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'style-failure.json') -Encoding utf8;throw}
