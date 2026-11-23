from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-split-preferences-20261124');dst=Path('docs/testing/results/2026-11-24-split-preferences-evidence');dst.mkdir(exist_ok=True)
names=['config-red.log','config-first-green-fixture-error.log','config-green.log','ui-red.log','ui-green.log','drag-green.log','serialization.log','rollback-red.log','rollback-green.log','source-alignment-red.log','source-alignment-green.log','noop-red.log','noop-green.log','broad-first.log','broad-final.log','checks.json','checks-before-noop.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log','binary-hashes.json','binary-hashes-before-noop.json','source-fingerprints.json','upstream-source.json','upstream-reference.log','upstream-states.json','upstream-overlay.json','upstream-split-reference_test.go','unused-upstream-tests.go','legacy-reader-source.json','legacy-reader.log','legacy-config/ui.go','legacy-config/ui_groups.go','legacy-config/legacy_test.go','checks.py','build.py','archive.py','verify-split-preferences.py','windows-status.json','local-environment.json']
entries=[]
for name in names:
 p=src/name;raw=p.read_bytes();value=raw.decode('utf-8-sig')
 if ('/'+'Users'+'/') in value or ('C:'+chr(92)+'Users'+chr(92)) in value:raise SystemExit('Private path in '+name)
 rel=Path('local')/(name+'.txt' if name.endswith('.go') else name);q=dst/rel;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(value)
 entries.append(dict(path=str(rel),sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest()))
(dst/'manifest.json').write_text(json.dumps({'schema_version':1,'scope':'local product checks, pinned source contract and actual prior configuration reader only; Windows not_run','artifacts':entries},indent=2)+'\n')
print('Archived',len(entries),'whitelisted artifacts')
