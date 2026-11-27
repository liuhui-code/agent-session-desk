$ErrorActionPreference='Stop';$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$o=Join-Path $b 'glyph-probe3';New-Item -ItemType Directory $o|Out-Null;$r=@{result='started'};$w=[IntPtr]::Zero
try{
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $b 'desktop-api.ps1'))))
Add-Type -AssemblyName System.Drawing
$h=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json
$before=[DeskFlow]::GetForegroundWindow()
if(@([DeskFlow]::Find('ASD Glyph Probe3 20261127')).Count){throw 'Fresh title'}
$code="& ([ScriptBlock]::Create([IO.File]::ReadAllText('"+(Join-Path $b 'glyph-child3.ps1').Replace("'","''")+"')))"
$a='-w new new-tab --title "ASD Glyph Probe3 20261127" --suppressApplicationTitle "'+$h.shell.path+'" -NoProfile -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
$si=[Diagnostics.ProcessStartInfo]::new();$si.FileName=$h.terminal.path;$si.Arguments=$a;$null=[Diagnostics.Process]::Start($si)
$end=[DateTime]::UtcNow.AddSeconds(20);do{$ws=@([DeskFlow]::Find('ASD Glyph Probe3 20261127'));if($ws.Count-eq 1){$w=$ws[0];break};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow-lt $end)
if($w-eq [IntPtr]::Zero){throw 'No private window'}
$end=[DateTime]::UtcNow.AddSeconds(40);while(!(Test-Path (Join-Path $o 'ready'))-and [DateTime]::UtcNow-lt $end){Start-Sleep -Milliseconds 200};if(!(Test-Path (Join-Path $o 'ready'))){throw 'Child not ready'}
$null=[DeskFlow]::SetForegroundWindow($w);Start-Sleep -Seconds 2
if([DeskFlow]::GetForegroundWindow()-ne $w){throw 'Foreground guard'}
$rc=[DeskFlow+R]::new();if(![DeskFlow]::GetWindowRect($w,[ref]$rc)){throw 'Rect'}
$im=[Drawing.Bitmap]::new($rc.Right-$rc.Left,$rc.Bottom-$rc.Top);$gr=[Drawing.Graphics]::FromImage($im);try{$gr.CopyFromScreen($rc.Left,$rc.Top,0,0,$im.Size);$im.Save((Join-Path $o 'glyphs.png'))}finally{$gr.Dispose();$im.Dispose()}
$r.result='captured'
} catch{$r.result='failed';$r.error=$_.Exception.Message} finally{if($w-ne [IntPtr]::Zero){$null=[DeskFlow]::SendMessage($w,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)};if($before-ne [IntPtr]::Zero){$null=[DeskFlow]::SetForegroundWindow($before)}}
$r|ConvertTo-Json|Set-Content (Join-Path $o 'r')
