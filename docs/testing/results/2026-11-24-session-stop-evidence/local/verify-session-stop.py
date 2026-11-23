from pathlib import Path
import json,hashlib,re
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
if root.name!='agent-session-desk':raise SystemExit('Run archived verifier from the repository')
e=root/'docs/testing/results/2026-11-24-session-stop-evidence';local=e/'local'
manifest=json.loads((e/'manifest.json').read_text())['artifacts']
for a in manifest:
 p=e/a['path'];assert p.is_file() and hashlib.sha256(p.read_bytes()).hexdigest()==a['sha256'],a['path']
fps=json.loads((local/'source-fingerprints.json').read_text())
actual_sources={str(p.relative_to(root)) for p in root.rglob('*.go')}|{'go.mod','go.sum'}
assert set(fps)==actual_sources
for name,sha in fps.items():assert hashlib.sha256((root/name).read_bytes()).hexdigest()==sha,name
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
status=json.loads((local/'windows-status.json').read_text());assert status['status']=='not_run' and not status['upload_started'] and not status['native_test_started'] and not status['console_flow_started']
source=json.loads((local/'source-reference.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and source['test_overlay_disables_ui_tests']==456
states=json.loads((local/'upstream-states.json').read_text());assert [s['focused'] for s in states]==[1,0,1,0,1] and all(s['target_id']=='original-id' for s in states)
assert 'PASS' in (local/'upstream-reference-run2.log').read_text()
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==3 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no stop confirmation' in (local/'ui-red.log').read_text() and 'no reviewed Session stop' in (local/'manager-red.log').read_text()
assert 'silently hid' in (local/'field-review-red.log').read_text() and 'ok ' in (local/'field-review-green.log').read_text()
expected={a['path'] for a in manifest}|{'manifest.json','verification.json','.gitattributes'}
actual={str(p.relative_to(e)) for p in e.rglob('*') if p.is_file()};assert actual<=expected,actual-expected
result=dict(result='passed_local_evidence_consistency',artifacts=len(manifest),source_fingerprints=len(fps),upstream_confirmation_states=len(states),checks=len(checks),windows=status['status'],scope='does not execute product tests or establish desktop/native/agent/resource/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
