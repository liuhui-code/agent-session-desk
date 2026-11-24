from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-launch-env-20261124');dst=Path('docs/testing/results/2026-11-24-launch-environment-evidence');dst.mkdir(exist_ok=True)
names=['env-red.log','env-green.log','env-error-red.log','env-error-green.log','env-location-red.log','env-location-green.log','env-boundaries-first.log','env-layout.log','env-final.log','env-frames.json','environment.json','windows-status.json','unchanged-contracts.json','checks.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-render.log','catalog-check.log','source-fingerprints.json','build.py','binary-hashes.json','candidate-payload-check.json','archive.py','verify-launch-environment.py']
artifacts=[]
for n in names:
 raw=(src/n).read_bytes();old=raw.decode('utf-8');text=old.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,n
 rel='local/'+n;q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=text!=old))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='controlled UI service boundary and synthetic model frames; no real child/backend, terminal, desktop or native Windows execution',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n');print('Archived',len(artifacts),'explicitly whitelisted artifacts')
