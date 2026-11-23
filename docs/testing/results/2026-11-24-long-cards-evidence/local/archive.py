from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-long-cards-20261124');dst=Path('docs/testing/results/2026-11-24-long-cards-evidence');dst.mkdir(exist_ok=True)
names=['card-red.log','card-green.log','long-card-boundaries.log','ui-first.log','inert-basename-red.log','inert-basename-green.log','checks.json','checks-before-inert-basename.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log','binary-hashes.json','binary-hashes-before-inert-basename.json','source-fingerprints.json','local-environment.json','windows-status.json','card-cases.json','upstream-source.json','upstream-reference.log','upstream-reference-final.log','upstream-frames.json','upstream-overlay.json','upstream-card-frames_test.go','unused-upstream-tests.go','before-production.json','workspace_sidebar.go.before','session_title.go.before','product-before-overlay.json','product-after-overlay.json','product-before.json','product-after.json','product-before.log','product-after.log','product-after-final.log','product-card-frames_test.go','comparison.json','compare-frames.py','checks.py','build.py','archive.py','verify-long-cards.py']
entries=[]
for name in names:
 p=src/name;raw=p.read_bytes();value=raw.decode('utf-8-sig').replace(str(Path.cwd()),'<WORKSPACE>')
 if ('/'+'Users'+'/') in value or ('C:'+chr(92)+'Users'+chr(92)) in value:raise SystemExit('Unexpected private path in '+name)
 rel=Path('local')/(name+'.txt' if name.endswith(('.go','.go.before')) else name);q=dst/rel;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(value)
 entries.append(dict(path=str(rel),sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(value!=raw.decode('utf-8-sig'))))
(dst/'manifest.json').write_text(json.dumps({'schema_version':1,'scope':'local product/model and pinned source text-contract evidence only; Windows not_run','artifacts':entries},indent=2)+'\n')
print('Archived',len(entries),'explicit artifacts')
