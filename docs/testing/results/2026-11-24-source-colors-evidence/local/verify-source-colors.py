from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-source-colors-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==38 and len({v['path'] for v in a})==38
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['ui-red.log','affected-first.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['ui-green.log','affected-green.log','color-regressions.log','source-reference.log','product-reference.log']:assert 'ok ' in (local/n).read_text(),n
log=(local/'color-regressions.log').read_text();assert 'PASS: TestSourceColorAdaptationLeavesNativeRGBAndIndexedAttributesExact' in log and log.count('--- PASS: TestWorkspaceUsesPinnedSourceEmittedRGBWithoutChangingAuthoredPalette/')==2
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('48d44a8')
for n in ['go.mod','go.sum']:assert subprocess.check_output(['git','show',baseline+':'+n],cwd=root)==(root/n).read_bytes()
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and source['source_helper_executed'] and not source['source_ui_executed'] and source['disabled_ui_tests']==456
assert hashlib.sha256((root/'third_party/agent-deck/LICENSE').read_bytes()).hexdigest()==source['sha256']['LICENSE']
policy=json.loads((local/'reference-policy.json').read_text());assert not policy['HOME_changed'] and policy['environment_allowlist'] and policy['network_outbound_denied'] and policy['real_home_file_access_denied'] and policy['source_worktree_after']=='clean'
s=json.loads((local/'source-reference.json').read_text());p=json.loads((local/'product-reference.json').read_text());assert len(s['frames'])==len(p['frames'])==144 and s['normalization']==p['normalization'] and len(s['palette'])==len(p['palette'])==20
r=json.loads((local/'role-comparison.json').read_text());assert len(r['roles'])==6 and r['matching_roles']==6 and all(v['matches'] for v in r['roles'])
assert len(r['palette'])==r['matching_palette_roles']==20 and all(v['matches'] for v in r['palette'])
refs=json.loads((local/'conversion-sources.json').read_text());assert len(refs)==4 and all(v['role']=='diagnosis_only_no_implementation_copied' and v['license']=='MIT' for v in refs)
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_clipboard_called'] and not windows['system_settings_written']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==2 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,source_component_frames=144,product_component_frames=144,exact_authored_and_display_roles=20,exact_combined_style_roles=6,native_rgb_and_indexed_preserved=True,dependency_files_unchanged=True,windows=windows['status'],scope='not full source UI, 16/256-color profiles, native Windows, real agent or complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
