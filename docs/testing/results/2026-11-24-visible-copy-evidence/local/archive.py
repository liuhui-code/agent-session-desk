from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-visible-copy-20261124');dst=Path('docs/testing/results/2026-11-24-visible-copy-evidence');dst.mkdir(exist_ok=True)
names=['backend-red.log','backend-green.log','ui-red.log','ui-green.log','backend-manager.log','queued-target-red.log','queued-target-green.log','command-binding-red.log','command-binding-green.log','token-guard-negative.log','tea-output.log','running-cancellation.log','focused-final.log','C1-red.log','C1-green.log','encoding-green.log','go-test-before-encoding-fix.log','checks-before-encoding-fix.json','binary-hashes-before-encoding-fix.json','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log','binary-hashes.json','source-fingerprints.json','local-environment.json','windows-status.json','pinned-framework-source.json','context7-resolve.json','context7-v2-docs.json','context7-main-docs.json','normalization-cases.json','upstream-source.json','upstream-reference.log','upstream-normalization.json','upstream-overlay.json','upstream-visible-copy_test.go','unused-upstream-tests.go','product-overlay.json','product-visible-copy_test.go','product-normalization.log','product-normalization-final.log','product-normalization.json','normalization-comparison.json','token-negative-overlay.json','without-request-token.go','checks.py','build.py','archive.py','verify-visible-copy.py']
artifacts=[]
for name in names:
 p=src/name;raw=p.read_bytes();escaped=False
 try:decoded=raw.decode('utf-8-sig')
 except UnicodeDecodeError:decoded=raw.decode('utf-8-sig',errors='backslashreplace');escaped=True
 value=decoded.replace(str(Path.cwd()),'<WORKSPACE>')
 if ('/'+'Users'+'/') in value or ('C:'+chr(92)+'Users'+chr(92)) in value:raise SystemExit('Unexpected private path in '+name)
 rel=Path('local')/(name+'.txt' if name.endswith('.go') else name);q=dst/rel;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(value)
 artifacts.append(dict(path=str(rel),sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(decoded!=value),invalid_utf8_escaped=escaped))
(dst/'manifest.json').write_text(json.dumps({'schema_version':1,'scope':'local subprocess contracts, source normalization with injected callbacks and Tea output to memory; Windows not_run','artifacts':artifacts},indent=2)+'\n')
print('Archived',len(artifacts),'whitelisted artifacts')
