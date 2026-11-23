from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-themes-20261124');dst=Path('docs/testing/results/2026-11-24-themes-evidence');dst.mkdir(exist_ok=True)
names=['config-red.log','config-green.log','config-boundaries.log','ui-red.log','ui-green.log','affected.log','theme-regressions.log','product-reference.log','source-reference.log','old-reader.log','upstream-palette.json','product-palette.json','palette-comparison.json','upstream-source.json','reference-policy.json','environment.json','windows-status.json','framework-sources.json','source-fingerprints.json','binary-hashes.json','context7-resolve.json','context7-docs.json','source-overlay.json','product-overlay.json','old-reader-overlay.json','source-theme_test.go','product-theme_test.go','old-ui.go','empty-config-test.go','old-reader_test.go','unused-upstream-tests.go','run-reference.py','checks.py','build.py','archive.py','verify-themes.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();original=raw.decode('utf-8');text=original.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name+('.txt' if name.endswith('.go') else '');q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(text!=original)))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='local model/config, isolated source palette and old-code overlay only; Windows not_run',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n')
print('Archived',len(artifacts),'explicitly whitelisted artifacts')
