from pathlib import Path
import hashlib,json,shutil
src=Path('/private/tmp/asd-session-stop-20261124');dst=Path('docs/testing/results/2026-11-24-session-stop-evidence');dst.mkdir(exist_ok=True)
(src/'windows-status.json').write_text(json.dumps({'status':'not_run','upload_started':False,'native_test_started':False,'console_flow_started':False,'reason':'Automatic approval rejected scp: destination and binary payload authorization insufficient; explicit user approval requested and pending. No alternate transport attempted.','destination':'windows-pc:asd-acceptance-20261122/','planned_payload':['asd-session-stop.exe','session-stop-manager.exe','session-stop-ui.exe'],'scope':'local cross-builds and prepared scripts are not native Windows evidence'},indent=2)+'\n')
manifest=[]
names=['manager-red.log','manager-green.log','ui-red.log','ui-green.log','ui-controls.log','focused.log','focused-final.log','field-review-red.log','field-review-green.log','checks.json','checks-before-extra-cases.json','checks-before-field-review.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log','binary-hashes.json','binary-hashes-before-extra-cases.json','binary-hashes-before-field-review.json','source-fingerprints.json','source-fingerprints-before-field-review.json','source-reference.json','upstream-reference.log','upstream-reference-run2.log','upstream-states.json','upstream-overlay.json','unused-upstream-tests.go','upstream-stop-reference_test.go','checks.py','build.py','make-native.py','archive.py','verify-session-stop.py','run-session-stop-tests.ps1','run-session-stop-console.ps1','windows-status.json']
for name in names:
 p=src/name
 if not p.exists():raise SystemExit('Missing '+name)
 raw=p.read_bytes();text=raw.decode('utf-8-sig')
 # All artifacts are generated synthetic fixtures; prohibit user paths anyway.
 if ('/'+'Users'+'/') in text or 'C:\\Users\\' in text:raise SystemExit('Unexpected private path in '+name)
 rel=Path('local')/(name+'.txt' if name.endswith('.go') else name);target=dst/rel;target.parent.mkdir(exist_ok=True)
 target.write_text(text)
 manifest.append(dict(path=str(rel),sha256=hashlib.sha256(target.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest()))
(dst/'manifest.json').write_text(json.dumps({'schema_version':1,'scope':'local evidence only; no native Windows execution this slice','artifacts':manifest},indent=2)+'\n')
print('Archived',len(manifest),'explicit artifacts')
