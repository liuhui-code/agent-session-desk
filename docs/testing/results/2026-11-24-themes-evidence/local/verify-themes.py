from pathlib import Path
import hashlib,json,re,subprocess
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk'
e=root/'docs/testing/results/2026-11-24-themes-evidence';local=e/'local'
a=json.loads((e/'manifest.json').read_text())['artifacts'];assert len(a)==43 and len({v['path'] for v in a})==43
for v in a:assert hashlib.sha256((e/v['path']).read_bytes()).hexdigest()==v['sha256'],v['path']
expected={v['path'] for v in a}|{'manifest.json','verification.json','.gitattributes'}
actual={str(f.relative_to(e)) for f in e.rglob('*') if f.is_file()};assert actual<=expected
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(f.relative_to(root)) for f in root.rglob('*.go')}|{'go.mod','go.sum'}
for n,h in fp.items():assert hashlib.sha256((root/n).read_bytes()).hexdigest()==h,n
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for n in ['config-red.log','ui-red.log']:assert 'FAIL' in (local/n).read_text(),n
for n in ['config-green.log','config-boundaries.log','ui-green.log','affected.log','theme-regressions.log','product-reference.log']:assert 'ok ' in (local/n).read_text(),n
assert 'PASS: TestASDThemePaletteReference' in (local/'source-reference.log').read_text()
assert 'PASS: TestASDOldUIReaderRejectsV5' in (local/'old-reader.log').read_text()
source=json.loads((local/'upstream-source.json').read_text());assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not source['worktree'] and not source['production_source_modified'] and source['disabled_ui_tests']==456
policy=json.loads((local/'reference-policy.json').read_text());assert policy['network_outbound_denied'] and policy['real_home_file_access_denied'] and not policy['HOME_changed'] and policy['source_worktree_after']=='clean'
s=json.loads((local/'upstream-palette.json').read_text());p=json.loads((local/'product-palette.json').read_text());assert s==p and len(s)==2 and all(len(v['colors'])==10 for v in s)
c=json.loads((local/'palette-comparison.json').read_text())['cases'];assert len(c)==20 and all(v['matches'] and v['source']==v['product'] for v in c)
baseline=json.loads((local/'environment.json').read_text())['baseline'];assert subprocess.check_output(['git','show',baseline+':internal/config/ui.go'],cwd=root,text=True)==(local/'old-ui.go.txt').read_text()
windows=json.loads((local/'windows-status.json').read_text());assert windows['status']=='not_run' and not windows['upload_attempted'] and not windows['native_executed'] and not windows['desktop_executed'] and not windows['system_clipboard_called']
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==4 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
result=dict(result='passed_local_evidence_consistency',artifacts=len(a),source_fingerprints=len(fp),checks=5,palette_values=20,source_tests_excluded=456,old_reader_writer_rejection_verified=True,windows=windows['status'],scope='not full source UI, System theme, native Windows, clipboard, real program or visual/release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
