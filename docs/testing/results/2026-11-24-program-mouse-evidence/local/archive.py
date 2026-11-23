from pathlib import Path
import hashlib,json
src=Path('/private/tmp/asd-program-mouse-20261124');dst=Path('docs/testing/results/2026-11-24-program-mouse-evidence');dst.mkdir(exist_ok=True)
names=['program-red.log','program-green.log','release-red.log','release-green.log','growth-red.log','growth-green.log','lifecycle.log','lifecycle-final.log','reserve-negative.log','without-release-reserve.go','negative-overlay.json','negative-result.json','protocol-sources.json','windows-status.json','environment.json','binary-hashes.json','source-fingerprints.json','catalog-check.log','checks.py','build.py','archive.py','verify-program-mouse.py','checks.json','go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log']
artifacts=[]
for name in names:
 raw=(src/name).read_bytes();original=raw.decode('utf-8');text=original.replace(str(Path.cwd()),'<WORKSPACE>');assert '/'+'Users'+'/' not in text and 'C:'+chr(92)+'Users'+chr(92) not in text,name
 rel='local/'+name+('.txt' if name.endswith('.go') else '');q=dst/rel;q.parent.mkdir(exist_ok=True);q.write_text(text)
 artifacts.append(dict(path=rel,sha256=hashlib.sha256(q.read_bytes()).hexdigest(),original_sha256=hashlib.sha256(raw).hexdigest(),workspace_path_normalized=(text!=original)))
(dst/'manifest.json').write_text(json.dumps(dict(schema_version=1,scope='local Go model-service and existing memory-pipe contracts only; Windows not_run',artifacts=artifacts),indent=2)+'\n')
(dst/'.gitattributes').write_text('*.log -text -whitespace\n')
print('Archived',len(artifacts),'explicitly whitelisted artifacts')
