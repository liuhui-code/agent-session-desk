from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-source-colors-20261124');dst=Path('docs/testing/results/2026-11-24-source-colors-evidence');dst.mkdir(exist_ok=True)
names=['ui-red.log','ui-green.log','affected-first.log','affected-green.log','color-regressions.log','source-reference.log','product-reference.log','source-reference.json','product-reference.json','role-comparison.json','compare-reference.py','source-reference_test.go','product-reference_test.go','source-overlay.json','product-overlay.json','run-reference.py','unused-upstream-tests.go','upstream-source.json','reference-policy.json','conversion-sources.json','context7-resolve.json','context7-docs.json','environment.json','windows-status.json','checks.py','build.py','archive.py','verify-source-colors.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-render.log','catalog-check.log','source-fingerprints.json','binary-hashes.json']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();original=raw.decode('utf-8');text=original.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name+('.txt' if name.endswith('.go') else '');q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(text!=original)))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='local rendered UI/RGB models and sandboxed pinned-source helper; exact TrueColor roles only, native Windows/desktop not_run',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n');print('Archived',len(artifacts),'explicitly whitelisted artifacts')
