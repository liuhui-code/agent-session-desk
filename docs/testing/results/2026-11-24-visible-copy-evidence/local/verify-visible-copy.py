from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk','Run archived verifier in repository'
e=root/'docs/testing/results/2026-11-24-visible-copy-evidence';local=e/'local';artifacts=json.loads((e/'manifest.json').read_text())['artifacts']
for a in artifacts:assert hashlib.sha256((e/a['path']).read_bytes()).hexdigest()==a['sha256'],a['path']
assert sum(a['invalid_utf8_escaped'] for a in artifacts)==2
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(p.relative_to(root)) for p in root.rglob('*.go')}|{'go.mod','go.sum'}
for name,sha in fp.items():assert hashlib.sha256((root/name).read_bytes()).hexdigest()==sha,name
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
first=json.loads((local/'checks-before-encoding-fix.json').read_text());assert first[0]['check']=='go-test' and first[0]['exit_code']==1
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
status=json.loads((local/'windows-status.json').read_text());assert status['status']=='not_run' and not status['upload_attempted_this_slice'] and not status['native_test_started'] and not status['desktop_started'] and not status['real_system_clipboard_validation']
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and not source['production_source_modified'] and source['disabled_ui_tests']==456
s=json.loads((local/'upstream-normalization.json').read_text());p=json.loads((local/'product-normalization.json').read_text());c=json.loads((local/'normalization-comparison.json').read_text())['cases'];assert len(s)==len(p)==len(c)==10
for a,b,x in zip(s,p,c):assert a['name']==b['name']==x['name'] and a['text']==b['text'] and x['text_matches'] and b['capture_encoding_valid'] and a['capture_calls']==1
assert sum(x['clipboard_callback_calls'] for x in s)==9
assert 'PASS: TestASDVisibleCopyReference' in (local/'upstream-reference.log').read_text()
for name in ['backend-red.log','ui-red.log','queued-target-red.log','command-binding-red.log','C1-red.log','C1-green.log','token-guard-negative.log']:assert 'FAIL' in (local/name).read_text(),name
for name in ['backend-green.log','ui-green.log','queued-target-green.log','command-binding-green.log','encoding-green.log','tea-output.log','running-cancellation.log']:assert 'ok ' in (local/name).read_text(),name
clip=(root/'internal/ui/workspace_clipboard.go').read_text();negative=(local/'without-request-token.go.txt').read_text();assert negative==clip.replace('if current == nil || msg.token != current.token {','if current == nil {',1)
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==5 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
expected={a['path'] for a in artifacts}|{'manifest.json','verification.json','.gitattributes'};actual={str(p.relative_to(e)) for p in e.rglob('*') if p.is_file()};assert actual<=expected,actual-expected
result=dict(result='passed_local_evidence_consistency',artifacts=len(artifacts),source_fingerprints=len(fp),normalization_cases=len(s),source_injected_clipboard_callbacks=9,checks=len(checks),prior_full_failure_retained=True,escaped_diagnostic_logs=2,windows=status['status'],scope='not native pane execution, real clipboard/desktop, physical modifiers, agent, resource or release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
