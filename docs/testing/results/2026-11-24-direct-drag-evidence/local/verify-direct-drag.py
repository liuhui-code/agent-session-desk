from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-direct-drag-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==25 and len({v['path'] for v in a})==25
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['terminal-red.log','ui-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['terminal-green.log','ui-green.log','lifecycle-green.log','lifecycle-final.log','capability.log']:assert 'ok ' in (local/n).read_text(),n
source=json.loads((local/'protocol-sources.json').read_text());assert source['commit']=='20aa1d7254ee93f64dafa059ef89341d9767470a' and source['second_local_snapshot_matches'] and len(source['files'])==6
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_clipboard_called']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==3 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,protocol_source_files=6,windows=windows['status'],scope='not live psmux/ConPTY, real mouse, system clipboard or full experience acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
