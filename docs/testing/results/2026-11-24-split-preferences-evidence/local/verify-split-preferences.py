from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk','Run archived verifier from repository'
e=root/'docs/testing/results/2026-11-24-split-preferences-evidence';local=e/'local';artifacts=json.loads((e/'manifest.json').read_text())['artifacts']
for a in artifacts:assert hashlib.sha256((e/a['path']).read_bytes()).hexdigest()==a['sha256'],a['path']
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(p.relative_to(root)) for p in root.rglob('*.go')}|{'go.mod','go.sum'}
for name,sha in fp.items():assert hashlib.sha256((root/name).read_bytes()).hexdigest()==sha,name
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
status=json.loads((local/'windows-status.json').read_text());assert status['status']=='not_run' and not status['upload_attempted_this_slice'] and not status['native_test_started'] and not status['desktop_started']
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and not source['production_source_modified'] and source['disabled_ui_test_files']==456
states=json.loads((local/'upstream-states.json').read_text());assert [(s['preview_pct'],s['disk_preview_pct']) for s in states]==[(65,65),(60,60),(50,60),(50,50),(90,50),(90,90),(90,90)]
for a,b in [(1,2),(3,4),(5,6)]:assert states[a]['file_sha256']==states[b]['file_sha256']
for a,b in [(0,1),(2,3),(4,5)]:assert states[a]['file_sha256']!=states[b]['file_sha256']
assert 'PASS: TestASDPersistentSplitReference' in (local/'upstream-reference.log').read_text()
old=json.loads((local/'legacy-reader-source.json').read_text())
for name,sha in old['unchanged_source_sha256'].items():assert hashlib.sha256((local/'legacy-config'/(name+'.txt')).read_bytes()).hexdigest()==sha
assert 'PASS: TestOldReaderRefusesV4WithoutChangingBytes' in (local/'legacy-reader.log').read_text()
for name in ['config-red.log','ui-red.log','rollback-red.log','source-alignment-red.log','noop-red.log']:assert 'FAIL' in (local/name).read_text(),name
for name in ['config-green.log','rollback-green.log','source-alignment-green.log','noop-green.log']:assert 'ok ' in (local/name).read_text(),name
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==4 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
expected={a['path'] for a in artifacts}|{'manifest.json','verification.json','.gitattributes'};actual={str(p.relative_to(e)) for p in e.rglob('*') if p.is_file()};assert actual<=expected,actual-expected
result=dict(result='passed_local_evidence_consistency',artifacts=len(artifacts),source_fingerprints=len(fp),upstream_states=len(states),checks=len(checks),old_reader='passed_on_local_host',windows=status['status'],scope='not native, desktop, physical mouse/IME, agent, resource or release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
