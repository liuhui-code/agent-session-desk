from pathlib import Path
import hashlib,json,subprocess
old=Path('/private/tmp/asd-sidebar-desktop-20261123');out=Path('/private/tmp/asd-groups-desktop-20261123')
product=out/'asd-group-desktop.exe';sha=hashlib.sha256(product.read_bytes()).hexdigest()
subs=[('sidebar-navigation-desktop-20261123','group-workspace-desktop-20261123'),('asd-sidebar-navigation','asd-group-desktop'),('AsdSidebarNavigationSetup20261123','AsdGroupWorkspaceSetup20261123'),('AsdSidebarNavigationDesktop20261123','AsdGroupWorkspaceDesktop20261123'),('sidebar-setup-outcome','groups-setup-outcome'),('setup-sidebar-desktop.ps1','setup-groups-desktop.ps1'),('sidebar-desktop-driver.ps1','groups-desktop-driver.ps1'),('f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277',sha)]
def adapt(s):
 for a,b in subs:s=s.replace(a,b)
 return s
setup=adapt((old/'setup-sidebar-desktop.ps1').read_text())
setup=setup.replace("$promptScript=Join-Path $work 'prompt.ps1'", "foreach($dir in @('root-default','review-default')){New-Item -ItemType Directory (Join-Path $work $dir)|Out-Null}\n$promptScript=Join-Path $work 'prompt.ps1'")
setup=setup.replace("function global:prompt", "$identityRoot=Split-Path -Parent $PSCommandPath\nfunction global:prompt")
setup=setup.replace("(Join-Path $PWD ($label.Replace(' ','_')+'.history.txt'))", "(Join-Path $identityRoot ($label.Replace(' ','_')+'.history.txt'))")
setup=setup.replace("path=$p.Path}|ConvertTo-Json|Set-Content ($label.Replace(' ','_')+'.identity.json')", "path=$p.Path;cwd=$PWD.Path}|ConvertTo-Json|Set-Content (Join-Path $identityRoot ($label.Replace(' ','_')+'.identity.json'))")
assert 'cwd=$PWD.Path' in setup
(out/'setup-groups-desktop.ps1').write_text(setup)
(out/'start-groups-setup.ps1').write_text(adapt((old/'start-sidebar-setup.ps1').read_text()))
start=adapt((old/'start-sidebar-desktop.ps1').read_text()).replace('-Minutes 22','-Minutes 42')
(out/'start-groups-desktop.ps1').write_text(start)
driver=adapt((old/'sidebar-desktop-driver.ps1').read_text()).replace('AddMinutes(20)','AddMinutes(40)')
driver=driver.replace("@('Copy A','Copy B','我的会话')", "@('Copy A','Copy B','Group C','我的会话','项目 甲','审核')")
driver=driver.replace("$c.title -eq '我的会话'", "$c.title -in @('我的会话','项目 甲','审核')")
driver=driver.replace("$lines[$i] -match '[▾▸] 我的会话'", "$lines[$i] -match ('[▾▸] '+[regex]::Escape($c.title)+'(?:\\s|$)')")
driver=driver.replace("$c.title -ne '我的会话'", "$c.title -notin @('我的会话','项目 甲','审核')")
driver=driver.replace("return @{pid=$p.Id;", "return @{caps_lock=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock);pid=$p.Id;")
driver=driver.replace("$savedClipboard=$null;", "$capsBefore=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock)\n$savedClipboard=$null;")
driver=driver.replace("@{utc=[DateTime]::UtcNow.ToString('o');restored=$restored;", "@{utc=[DateTime]::UtcNow.ToString('o');caps_before=$capsBefore;caps_after=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock);restored=$restored;")
(out/'groups-desktop-driver.ps1').write_text(driver)
header=adapt((old/'sidebar-flow-header.ps1').read_text())
(out/'groups-flow-header.ps1').write_text(header)
files=sorted([*Path('internal').rglob('*.go'),*Path('cmd').rglob('*.go'),Path('go.mod'),Path('go.sum')])
(out/'product-source-before.json').write_text(json.dumps({'commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'git_status':subprocess.check_output(['git','status','--porcelain'],text=True),'product_sha256':sha,'metadata_helper_sha256':hashlib.sha256((out/'read-group-metadata.exe').read_bytes()).hexdigest(),'files':[{'path':str(p),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}for p in files]},indent=2)+'\n')
print(json.dumps({'product_sha256':sha,'source_files':len(files)}))
