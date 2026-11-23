from pathlib import Path
import hashlib,json,subprocess
old=Path('/private/tmp/asd-rename-20261123');out=Path('/private/tmp/asd-sidebar-desktop-20261123')
binary=out/'asd-sidebar-navigation.exe';product_hash=hashlib.sha256(binary.read_bytes()).hexdigest()
substitutions=[('session-rename-surface-20261123','sidebar-navigation-desktop-20261123'),('asd-rename-surface','asd-sidebar-navigation'),('AsdSessionRenameSurfaceSetup20261123','AsdSidebarNavigationSetup20261123'),('AsdSessionRenameSurfaceDesktop20261123','AsdSidebarNavigationDesktop20261123'),('rename-surface-setup-outcome','sidebar-setup-outcome'),('setup-rename-surface.ps1','setup-sidebar-desktop.ps1'),('rename-surface-driver.ps1','sidebar-desktop-driver.ps1'),('8c8aea2a17f94f593dd99a12f55d3f3c3e15db36b1201090243bab393fa573fd',product_hash)]
def adapt(text):
 for a,b in substitutions:text=text.replace(a,b)
 return text
setup=adapt((old/'setup-rename-surface.ps1').read_text())
setup=setup.replace("foreach($name in @('Copy A'))","foreach($name in @('Copy A','Copy B'))").replace("foreach($name in @('Copy_A'))","foreach($name in @('Copy_A','Copy_B'))")
setup=setup.replace("$backup=Join-Path $root 'settings.before'",'''$ui=Join-Path $root 'state/ui.toml'
[IO.File]::WriteAllText($ui,"schema_version = 1`nsidebar_density = `"compact`"`n",[Text.UTF8Encoding]::new($false))
[IO.File]::Copy($ui,(Join-Path $root 'ui-v1-before.toml'))
$backup=Join-Path $root 'settings.before' ''')
setup=setup.replace('go_tests_rerun=$true','go_tests_rerun=$false;desktop_only=$true')
setup=setup.replace("foreach($v in $expected)","$expected+=,@((Join-Path (Split-Path $backend) 'conpty.dll'),'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492')\n$expected+=,@((Join-Path (Split-Path $backend) 'OpenConsole.exe'),'b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775')\nforeach($v in $expected)")
(out/'setup-sidebar-desktop.ps1').write_text(setup)
(out/'start-sidebar-setup.ps1').write_text(adapt((old/'start-rename-surface-setup.ps1').read_text()))
(out/'start-sidebar-desktop.ps1').write_text(adapt((old/'start-rename-surface-desktop.ps1').read_text()))
driver=adapt((old/'rename-surface-driver.ps1').read_text())
driver=driver.replace('public static uint Click(){','public static uint Wheel(int delta){var a=new INPUT[1];a[0].data.mouse.flags=0x800;a[0].data.mouse.data=unchecked((uint)delta);return SendInput(1,a,Marshal.SizeOf(typeof(INPUT)));}\n public static uint Click(){')
driver=driver.replace("'clickListRow' {","'clickNavRow' {")
driver=driver.replace("@('Copy A','Copy B')){throw 'Unknown synthetic row title'}","@('Copy A','Copy B','我的会话')){throw 'Unknown synthetic row title'}")
a="for($i=0;$i -lt $lines.Length;$i++){if($lines[$i].StartsWith('  '+$c.title) -or $lines[$i].StartsWith('▶ '+$c.title) -or $lines[$i].StartsWith('┃ '+$c.title)){$row=$i;break}}"
b="for($i=2;$i -lt $lines.Length;$i++){if(($c.title -eq '我的会话' -and $lines[$i] -match '[▾▸] 我的会话') -or ($c.title -ne '我的会话' -and $lines[$i] -match ('^[ ▶┃]+'+[regex]::Escape($c.title)+'(?:\\s|$)'))){$row=$i;break}}"
assert a in driver;driver=driver.replace(a,b)
driver=driver.replace("!$lines[$row].StartsWith('  ╰ ') -and !$lines[$row].StartsWith('  │ ')","!$lines[$row].TrimStart().StartsWith('╰ ') -and !$lines[$row].TrimStart().StartsWith('│ ')")
a="for($i=0;$i -lt $count;$i++){Guard|Out-Null;[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;if([AsdDesktop]::Click()-ne 2){throw 'Mouse click failed'};if($i+1 -lt $count){Start-Sleep -Milliseconds 120}}"
b="if($c.wheel){if([int]$c.wheel -notin @(-120,120)){throw 'Invalid test wheel delta'};Guard|Out-Null;[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;if([AsdDesktop]::Wheel([int]$c.wheel)-ne 1){throw 'Mouse wheel failed'};$count=0}else{"+a+"}"
assert a in driver;driver=driver.replace(a,b)
(out/'sidebar-desktop-driver.ps1').write_text(driver)
header=adapt((old/'rename-surface-header.ps1').read_text()).replace("foreach($name in @('Copy_A'))","foreach($name in @('Copy_A','Copy_B'))")
(out/'sidebar-flow-header.ps1').write_text(header)
files=sorted([*Path('internal').rglob('*.go'),*Path('cmd').rglob('*.go'),Path('go.mod'),Path('go.sum')])
(out/'source-before.json').write_text(json.dumps({'commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'binary_sha256':product_hash,'files':[{'path':str(p),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in files]},indent=2)+'\n')
print(json.dumps({'product_sha256':product_hash,'go_source_files':len(files)}))
