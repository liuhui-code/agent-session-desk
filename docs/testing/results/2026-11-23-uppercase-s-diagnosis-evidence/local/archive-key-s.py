from pathlib import Path
import hashlib,json
scratch=Path('/private/tmp/asd-key-s-20261123')
root=Path('docs/testing/results/2026-11-23-uppercase-s-diagnosis-evidence')
files=[('native/'+str(p.relative_to(scratch/'native')).replace('\\','/'),p) for p in sorted((scratch/'native').rglob('*')) if p.is_file()]
local=[p for p in sorted(scratch.iterdir()) if p.is_file() and (p.suffix in ['.ps1','.py','.go','.json','.txt','.log'] or p.name in ['go.mod','go.sum']) and not p.name.endswith('-console.log')]
files += [('local/'+p.name+('.txt' if p.suffix=='.go' or p.name in ['go.mod','go.sum'] else ''),p) for p in local]
manifest=[]
for relative,path in files:
 raw=path.read_bytes()
 if path.suffix=='.png':data=raw;transform='Unmodified PNG; visually inspected. Label is capture intent, not an assertion of pixel/UIA synchronization.'
 else:
  text=raw.decode('utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig').replace('\r\n','\n').replace('\r','\n')
  for before,after in [(str(Path.cwd()),'<REPO>'),('<LOCAL_TEST_HOME>','<LOCAL_TEST_HOME>'),('<WINDOWS_TEST_USER>','<WINDOWS_TEST_USER>')]:text=text.replace(before,after)
  data=text.encode();transform='BOM decoded, UTF-8/LF, machine paths and test user redacted; Go helpers/modfiles stored as text.'
 target=root/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
 manifest.append({'path':relative,'original_bytes':len(raw),'original_sha256':hashlib.sha256(raw).hexdigest(),'archived_bytes':len(data),'archived_sha256':hashlib.sha256(data).hexdigest(),'transformation':transform})
(root/'manifest.json').write_text(json.dumps({'scope':'Native uppercase-S diagnostic only; no product patch, full keyboard/IME acceptance or release claim','native_zip_sha256':hashlib.sha256((scratch/'native-export.zip').read_bytes()).hexdigest(),'files':manifest},ensure_ascii=False,indent=2)+'\n')
print('Archived',len(manifest),'artifacts.')
