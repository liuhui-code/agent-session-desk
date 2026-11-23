from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-directory-pickers-evidence';local=e/'local'
artifacts=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(artifacts)==54 and len({x['path']for x in artifacts})==54
for v in artifacts:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path']for v in artifacts}|{'manifest.json','verification.json','.gitattributes'};assert {str(f.relative_to(e))for f in e.rglob('*')if f.is_file()}<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root))for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(v['exit_code']==0 for v in checks)
for n in ['recents-red.log','directories-red.log','group-picker-red.log','popup-colors-red.log','tiny-focus-red.log','explicit-tab-red.log','identity-red.log']:assert 'FAIL' in (local/n).read_text(),n
assert 'test timed out after 1m30s' in (local/'picker-first-timeout.log').read_text()
for n in ['recents-green.log','directories-green.log','picker-boundaries-green.log','popup-colors-green.log','affected-final.log','identity-green.log','picker-final.log','source-reference.log','product-reference.log']:assert 'ok ' in (local/n).read_text(),n
source=json.loads((local/'source-reference.json').read_text());frames=json.loads((local/'product-frames.json').read_text());assert len(source['frames'])==48 and len(frames)==128
styles=json.loads((local/'style-comparison.json').read_text());assert styles['cases']==styles['matching']==80 and all(v['matches'] for v in styles['results'])
semantics=json.loads((local/'semantic-comparison.json').read_text());assert len(semantics['cases'])==2 and all(v['matches']for v in semantics['cases'])
upstream=json.loads((local/'upstream-source.json').read_text());assert upstream['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not upstream['worktree'] and upstream['source_helper_executed'] and not upstream['source_ui_executed'] and upstream['disabled_ui_tests']==456
assert upstream['sha256']['LICENSE']==hashlib.sha256((root/'third_party/agent-deck/LICENSE').read_bytes()).hexdigest()
policy=json.loads((local/'reference-policy.json').read_text());assert not policy['HOME_changed'] and policy['environment_allowlist'] and policy['network_outbound_denied'] and policy['real_home_file_access_denied']
dep=json.loads((local/'dependency-sources.json').read_text());assert dep['version']=='v0.1.3' and dep['commit']=='8aadc77d46cf09f5dc529d41358a94a24516c0b6' and dep['license']=='MIT' and not dep['existing_module_versions_changed']
assert dep['sha256']['LICENSE']==hashlib.sha256((root/'third_party/sahilm-fuzzy/LICENSE').read_bytes()).hexdigest()
change=json.loads((local/'dependency-change.json').read_text());assert change['existing_versions_unchanged'] and len(change['added_sums'])==4 and not change['removed_sums']
assert (root/'go.mod').read_text().replace('\tgithub.com/sahilm/fuzzy v0.1.3\n','')==subprocess.check_output(['git','show',json.loads((local/'environment.json').read_text())['baseline']+':go.mod'],cwd=root,text=True)
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not any(windows[k]for k in ['upload_attempted','native_executed','desktop_executed','system_clipboard_called','system_settings_written'])
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==5 and all(re.fullmatch('[0-9a-f]{64}',v)for v in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(artifacts),source_fingerprints=len(fp),checks=5,source_component_states=48,product_frames=128,matching_style_cases=80,matching_semantic_cases=2,windows='not_run',scope='not full source UI, real terminal, native Windows, real agent or complete visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
