from pathlib import Path
import hashlib,json

scratch=Path('/private/tmp/asd-sidebar-desktop-20261123')
destination=Path('docs/testing/results/2026-11-23-sidebar-navigation-desktop-evidence')
files=[('native/'+str(p.relative_to(scratch/'native')).replace('\\','/'),p) for p in sorted((scratch/'native').rglob('*')) if p.is_file()]
local=['prepare-desktop.py','setup-sidebar-desktop.ps1','start-sidebar-setup.ps1','sidebar-desktop-driver.ps1','start-sidebar-desktop.ps1','sidebar-flow-header.ps1','sidebar-flow-body.ps1','sidebar-flow-body-corrected.ps1','sidebar-flow-body-final.ps1','sidebar-flow.ps1','sidebar-flow-resume.ps1','sidebar-flow-resume2.ps1','sidebar-flow-resume3.ps1','sidebar-flow-resume4.ps1','cleanup-export-sidebar.ps1','inspect-sidebar.ps1','source-before.json','source-after.json','setup-run.log','desktop-start.log','flow-run.log','flow-resume-run.log','flow-resume2-run.log','flow-resume3-run.log','flow-resume4-run.log','cleanup-run.log','inspect-first.log','inspect-second.log','inspect-third.log','inspect-fourth.log','catalog-check.log','source-limits.log','archive-desktop.py','verify-desktop.py']
files += [('local/'+name,scratch/name) for name in local]
manifest=[]
for relative,path in files:
 raw=path.read_bytes()
 if path.suffix.lower()=='.png':
  data=raw;transformation='Unmodified original PNG; visually inspected before archive.'
 else:
  encoding='utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig'
  text=raw.decode(encoding).replace('\r\n','\n').replace('\r','\n')
  for before,after in [(str(Path.cwd()),'<REPO>'),('<LOCAL_TEST_HOME>','<LOCAL_TEST_HOME>'),('<WINDOWS_TEST_USER>','<WINDOWS_TEST_USER>')]:text=text.replace(before,after)
  data=text.encode('utf-8');transformation='BOM decoded, UTF-8/LF, test user and machine paths redacted.'
 target=destination/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
 manifest.append({'path':relative,'original_bytes':len(raw),'original_sha256':hashlib.sha256(raw).hexdigest(),'archived_bytes':len(data),'archived_sha256':hashlib.sha256(data).hexdigest(),'transformation':transformation})
zip_file=scratch/'native-desktop-export.zip'
(destination/'manifest.json').write_text(json.dumps({'scope':'Bounded desktop navigation on explicit development terminal bundle; no model rerun, real agent or release claim','native_zip_sha256':hashlib.sha256(zip_file.read_bytes()).hexdigest(),'files':manifest},ensure_ascii=False,indent=2)+'\n')
print('Archived',len(manifest),'artifacts.')
