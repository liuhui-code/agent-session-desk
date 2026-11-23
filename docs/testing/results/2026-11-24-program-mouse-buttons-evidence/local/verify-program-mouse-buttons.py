from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-program-mouse-buttons-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==31 and len({x['path']for x in a})==31
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
assert {str(f.relative_to(e))for f in e.rglob('*')if f.is_file()}<=({v['path']for v in a}|{'manifest.json','verification.json','.gitattributes'})
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root))for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(v['exit_code']==0 for v in checks)
for n in ['wire-red.log','buttons-red.log','hover-red.log','hover-cancel-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['wire-green.log','buttons-green.log','hover-green.log','matrix-first.log','hover-cancel-green.log','lifecycle-first.log']:assert 'ok ' in (local/n).read_text(),n
protocol=json.loads((local/'protocol-sources.json').read_text());assert protocol['commit']=='20aa1d7254ee93f64dafa059ef89341d9767470a' and protocol['second_local_snapshot_matches'] and len(protocol['files'])==6 and len(protocol['limitations'])==4
assert protocol['files']['LICENSE']==hashlib.sha256((root/'third_party/psmux/LICENSE').read_bytes()).hexdigest()
framework=json.loads((local/'framework-sources.json').read_text());assert not framework['dependency_changes'] and len(framework['sha256'])==2
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('061f16d')
for n in ['go.mod','go.sum','internal/ui/design_tokens.go']:
 assert (root/n).read_bytes()==subprocess.check_output(['git','show',baseline+':'+n],cwd=root),n
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not any(windows[k]for k in ['upload_attempted','native_executed','desktop_executed','system_clipboard_called','system_settings_written'])
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==3 and all(re.fullmatch('[0-9a-f]{64}',v)for v in hashes.values())
payload=json.loads((local/'candidate-payload-check.json').read_text());assert len(payload)==3 and all(x['sha256']==hashes[x['name']] and x['format']=='PE amd64' and not x['native_executed'] for x in payload)
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,source_protocol_files=6,framework_modules=2,dependency_files_and_tokens_unchanged=True,windows='not_run',scope='not live psmux, target TUI, VT/Win32, physical mouse, real agent or complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
