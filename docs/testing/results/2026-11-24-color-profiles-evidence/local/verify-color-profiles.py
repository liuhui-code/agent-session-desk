from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-color-profiles-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==44 and len({v['path'] for v in a})==44
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['ui-red.log','defaults-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['ui-green.log','defaults-green.log','profile-final.log','source-reference.log','product-reference.log']:assert 'ok ' in (local/n).read_text(),n
log=(local/'profile-final.log').read_text();assert 'PASS: TestTeaColorCapabilityChangesEmitDefaultResetsAndKnownTrueColorRequests' in log and log.count('--- PASS: TestLimitedColorProgramsNeverRequestRGBDefaults/')==4
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert baseline.startswith('bffc6d7')
old_mod=subprocess.check_output(['git','show',baseline+':go.mod'],cwd=root,text=True);current=(root/'go.mod').read_text()
required=lambda s:dict(re.findall(r'^\s+(\S+) (v\S+)',s,re.M))
assert required(old_mod)==required(current) and subprocess.check_output(['git','show',baseline+':go.sum'],cwd=root)==(root/'go.sum').read_bytes()
dep=json.loads((local/'dependency-change.json').read_text());assert dep['module_versions_unchanged'] and dep['go_sum_unchanged'] and dep['root_required_modules']==required(current)
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and source['source_helper_executed'] and not source['source_ui_executed'] and source['disabled_ui_tests']==456
assert hashlib.sha256((root/'third_party/agent-deck/LICENSE').read_bytes()).hexdigest()==source['sha256']['LICENSE']
policy=json.loads((local/'reference-policy.json').read_text());assert not policy['HOME_changed'] and policy['environment_allowlist'] and policy['network_outbound_denied'] and policy['real_home_file_access_denied'] and policy['source_worktree_after']=='clean'
s=json.loads((local/'source-reference.json').read_text());p=json.loads((local/'product-reference.json').read_text());assert len(s)==len(p)==4 and all(len(v['labels'])==30 and 'ASD_PROFILE_DONE' in v['actual_program_output'] for v in s+p)
result=json.loads((local/'profile-comparison.json').read_text());assert len(result['profiles'])==4 and all(v['matches']==v['total']==30 and all(c['matches'] for c in v['cases']) for v in result['profiles'])
before=json.loads((local/'before-fix-profile-comparison.json').read_text());assert [v['matches'] for v in before['profiles']]==[30,23,11,0]
assert len(json.loads((local/'source-index-table.json').read_text()))==22
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_clipboard_called'] and not windows['system_settings_written']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==2 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
record=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,actual_renderer_profiles=4,matching_fixed_cases=120,matching_before=64,default_color_packet_tests=True,module_versions_unchanged=True,windows=windows['status'],scope='not real terminal palettes/default restoration, full source UI, native Windows, real agent or visual/release acceptance')
(e/'verification.json').write_text(json.dumps(record,indent=2)+'\n');print(json.dumps(record))
