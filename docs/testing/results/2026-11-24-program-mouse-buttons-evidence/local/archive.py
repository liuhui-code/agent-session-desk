from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-program-mouse-buttons-20261124');dst=Path('docs/testing/results/2026-11-24-program-mouse-buttons-evidence');dst.mkdir(exist_ok=True)
names=['wire-red.log','wire-green.log','buttons-red.log','buttons-green.log','hover-red.log','hover-green.log','matrix-first.log','hover-cancel-red.log','hover-cancel-green.log','lifecycle-first.log','protocol-sources.json','framework-sources.json','context7-resolve.json','context7-docs.json','environment.json','windows-status.json','checks.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-render.log','catalog-check.log','source-fingerprints.json','build.py','binary-hashes.json','candidate-payload-check.json','archive.py','verify-program-mouse-buttons.py']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();original=raw.decode('utf-8');text=original.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name;q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=text!=original))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='Go protocol memory-pipe and controlled UI model evidence; not real backend, terminal, desktop or native Windows',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n');print('Archived',len(artifacts),'explicitly whitelisted artifacts')
