$script:geometry=[Collections.Generic.List[object]]::new()
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Style driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$ids=@(CheckShells)
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '本地会话搜索';$null=WaitText '3 个结果';SearchGeometry 'final-empty-120' 120;$null=Shot 'search-final-empty-120-public'
 Paste '中文';$null=WaitText '› 中文会话';$null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '› 中文会话';SearchGeometry 'final-unicode-63' 63;$null=Shot 'search-final-unicode-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '› 中文会话';SearchGeometry 'final-unicode-49' 49;$null=Shot 'search-final-unicode-49-public'
 Key '{ESC}';$null=WaitText '管理模式';$null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';Text 'q';WaitExit
 $after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.Count -ne 3){throw 'Final style view changed sessions'}
 foreach($s in $after){if(@($sessions|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id}).Count -ne 1){throw 'Final style view changed original identity'}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$sessions;after=$after;shell_processes=@(CheckShells);ui_processes=$script:uis.ToArray();geometry=$script:geometry.ToArray();scope='Final placeholder index 240 and regular keyboard legend, same original Runs; preceding full interaction flow remains separately recorded'}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $testRoot 'style-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'style-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'style-failure.json') -Encoding utf8;throw}
