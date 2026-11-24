from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-launch-environment-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==28 and len({v['path']for v in a})==28
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
assert {str(f.relative_to(e))for f in e.rglob('*')if f.is_file()}<=({v['path']for v in a}|{'manifest.json','verification.json','.gitattributes'})
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root))for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(v['exit_code']==0 for v in checks)
for n in ['env-red.log','env-error-red.log','env-location-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['env-green.log','env-error-green.log','env-location-green.log','env-boundaries-first.log','env-layout.log','env-final.log']:assert 'ok ' in (local/n).read_text(),n
frames=json.loads((local/'env-frames.json').read_text());assert len(frames)==64 and len({(f['theme'],f['profile'],f['width'],f['height'])for f in frames})==64
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('82c6545')
contracts=json.loads((local/'unchanged-contracts.json').read_text());assert len(contracts)==7
for n,h in contracts.items():
 raw=(root/n).read_bytes();assert hashlib.sha256(raw).hexdigest()==h and raw==subprocess.check_output(['git','show',baseline+':'+n],cwd=root),n
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not any(windows[k]for k in ['upload_attempted','native_executed','desktop_executed','system_clipboard_called','system_settings_written'])
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==3 and all(re.fullmatch('[0-9a-f]{64}',v)for v in hashes.values())
payload=json.loads((local/'candidate-payload-check.json').read_text());assert len(payload)==3 and all(x['sha256']==hashes[x['name']] and x['format']=='PE amd64' and not x['native_executed']for x in payload)
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,model_frames=64,unchanged_contract_files=7,windows='not_run',scope='not real child environment, live backend, terminal, native Windows, agent or complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
