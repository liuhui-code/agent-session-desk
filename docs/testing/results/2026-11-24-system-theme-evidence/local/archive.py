from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-system-theme-20261124');dst=Path('docs/testing/results/2026-11-24-system-theme-evidence');dst.mkdir(exist_ok=True)
names=['config-red.log','config-green.log','appearance-red.log','appearance-green.log','appearance-unit.log','ui-red.log','ui-green.log','targeted.log','targeted-green.log','system-lifecycle.log','old-reader.log','old-reader_test.go','old-ui.go','old-reader-overlay.json','checks.py','build.py','archive.py','verify-system-theme.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-render.log','catalog-check.log','context7-resolve.json','context7-docs.json','context7-docs2.json','upstream-source.json','dark-mode-source.json','api-references.json','environment.json','windows-status.json','source-fingerprints.json','binary-hashes.json']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();original=raw.decode('utf-8');text=original.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name+('.txt' if name.endswith('.go') else '');q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(text!=original)))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='local appearance/config/UI models and baseline max-v5 reader/writer overlay; Windows native and desktop not_run',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n')
print('Archived',len(artifacts),'explicitly whitelisted artifacts')
