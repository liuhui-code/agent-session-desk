from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-direct-drag-20261124');dst=Path('docs/testing/results/2026-11-24-direct-drag-evidence');dst.mkdir(exist_ok=True)
names=['terminal-red.log','terminal-green.log','ui-red.log','ui-green.log','lifecycle-green.log','lifecycle-final.log','capability.log','context7-resolve.json','context7-docs.json','protocol-sources.json','windows-status.json','environment.json','binary-hashes.json','source-fingerprints.json','catalog-check.log','checks.py','build.py','archive.py','verify-direct-drag.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();text=raw.decode('utf-8');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name;q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_bytes(raw)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(raw).hexdigest()))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='local memory-pipe and Go model-service evidence only; Windows not_run',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n')
print('Archived',len(artifacts),'explicitly whitelisted artifacts')
