from pathlib import Path
import hashlib,json

scratch=Path('/private/tmp/asd-sidebar-navigation-20261123')
dest=Path('docs/testing/results/2026-11-23-sidebar-navigation-evidence')
logs=['red-config.log','green-config.log','red-grouped.log','green-grouped-first.log','red-mouse.log','green-mouse.log','regression-first.log','regression-second.log','regression-third.log','regression-fourth.log','quit-first-pass-input-masked.log','red-quit.log','green-quit.log','green-final.log','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-render.log','catalog-check.log','upstream-build.log','upstream-component-build.log','upstream-render.log','native-run.log']
files=[('local/'+n,scratch/n) for n in logs]
files += [('local/'+n,scratch/n) for n in ['product-source.json','upstream-source.json','binary-hashes.json','upstream-header-frames.json','upstream-reference_test.go','upstream-unused-tests.go','upstream-overlay.json','upstream-overlay-initial.json','upstream-testmain-reference.go','run-sidebar-navigation-tests.ps1','archive-navigation.py','verify-navigation.py']]
files += [('local/reference.sb',Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'))]
files += [('native/'+n,scratch/'native'/n) for n in ['sidebar-ui-tests.exe.log','sidebar-config-tests.exe.log','results.json']]
manifest=[]
for relative,source in files:
 if relative.endswith(".go"):
  relative += ".txt" # Evidence must not create another Go package under ./...
 raw=source.read_bytes()
 encoding='utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig'
 text=raw.decode(encoding).replace('\r\n','\n').replace('\r','\n')
 for before,after in [(str(Path.cwd()),'<REPO>'),('<LOCAL_TEST_HOME>','<LOCAL_TEST_HOME>'),('<WINDOWS_TEST_USER>','<WINDOWS_TEST_USER>')]:
  text=text.replace(before,after)
 data=text.encode('utf-8')
 target=dest/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
 manifest.append({'path':relative,'original_bytes':len(raw),'original_sha256':hashlib.sha256(raw).hexdigest(),'archived_bytes':len(data),'archived_sha256':hashlib.sha256(data).hexdigest(),'transformation':'Decoded BOM when present; UTF-8/LF; test user and machine paths redacted.'})
(dest/'manifest.json').write_text(json.dumps({'scope':'Local and Windows native model/config tests only; no current desktop execution','files':manifest},ensure_ascii=False,indent=2)+'\n')
