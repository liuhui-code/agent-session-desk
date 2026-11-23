from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-system-theme-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==36 and len({v['path'] for v in a})==36
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['config-red.log','ui-red.log','appearance-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['config-green.log','appearance-green.log','ui-green.log','targeted-green.log']:assert 'ok ' in (local/n).read_text(),n
log=(local/'system-lifecycle.log').read_text();assert log.count('--- PASS: TestThemeSettingsKeepFocusedChoiceInsideAllBoundedFrames/')==180
assert log.count('--- PASS: TestSystemTheme')==7
assert 'PASS: TestASDOldUIReaderRejectsV6' in (local/'old-reader.log').read_text()
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('1da0670')
assert subprocess.check_output(['git','show',baseline+':internal/config/ui.go'],cwd=root,text=True)==(local/'old-ui.go.txt').read_text()
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and not source['source_executed']
dark=json.loads((local/'dark-mode-source.json').read_text());assert dark['module_info']['Origin']['Hash']=='cfb25c8db9bff8e2e1f993e90f26326d4e68ed44'
assert hashlib.sha256((root/'third_party/dark-mode-go/LICENSE').read_bytes()).hexdigest()==dark['sha256']['LICENSE']
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_settings_written']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==5 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,synthetic_frames=180,system_model_tests=7,old_reader_writer_rejection_verified=True,windows=windows['status'],scope='not native Windows notifications/cleanup, source UI execution, real agent, complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
