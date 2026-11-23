from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-footer-modes-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==43 and len({v['path'] for v in a})==43
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['config-red.log','ui-red.log','forced-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['config-green.log','ui-green.log','forced-green.log','affected.log','footer-final.log','source-reference.log','product-reference.log']:assert 'ok ' in (local/n).read_text(),n
assert (local/'footer-final.log').read_text().count('--- PASS: TestWorkspaceFooterModesRetainBoundedFramesAndTerminalOwnership/')==112
assert 'PASS: TestASDOldUIReaderRejectsV7' in (local/'old-reader.log').read_text()
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('879d94e')
assert subprocess.check_output(['git','show',baseline+':internal/config/ui.go'],cwd=root,text=True)==(local/'old-ui.go.txt').read_text()
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and source['source_helper_executed'] and not source['source_ui_executed'] and source['disabled_ui_tests']==456
assert hashlib.sha256((root/'third_party/agent-deck/LICENSE').read_bytes()).hexdigest()==source['sha256']['LICENSE']
policy=json.loads((local/'reference-policy.json').read_text());assert not policy['HOME_changed'] and policy['environment_allowlist'] and policy['network_outbound_denied'] and policy['real_home_file_access_denied'] and policy['source_worktree_after']=='clean'
s=json.loads((local/'source-reference.json').read_text());p=json.loads((local/'product-reference.json').read_text());assert len(s['frames'])==len(p['frames'])==144 and s['normalization']==p['normalization'] and len(s['normalization'])==7
r=json.loads((local/'role-comparison.json').read_text());assert len(r['roles'])==6 and r['matching_roles']==3 and r['normalization_matches']
assert len(json.loads((local/'settings-inventory.json').read_text()))==40
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_clipboard_called']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==4 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,synthetic_frames=112,source_component_frames=144,product_component_frames=144,normalization_cases=7,exact_style_matches=3,exact_style_differences=3,source_settings_inventoried=40,old_reader_writer_rejection_verified=True,windows=windows['status'],scope='not full source UI, exact RGB equality, native Windows, real agent or complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
