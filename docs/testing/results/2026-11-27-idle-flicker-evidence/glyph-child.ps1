[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$Host.UI.RawUI.WindowTitle='ASD Glyph Probe3 20261127'
[Console]::Write(([char]27)+'[2J'+([char]27)+'[H')
[Console]::WriteLine('Synthetic glyph probe: large-circle / circle / bullet / half-circle / braille')
[Console]::Write([IO.File]::ReadAllText((Join-Path $b 'glyphs.ansi'),[Text.Encoding]::UTF8))
[Console]::Write(([char]27)+'[0m')
Set-Content (Join-Path $b 'glyph-probe3/ready') 'ready'
Start-Sleep -Seconds 45
