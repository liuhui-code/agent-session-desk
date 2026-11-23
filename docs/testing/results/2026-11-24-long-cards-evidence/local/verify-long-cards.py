from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve()
while root.name!='agent-session-desk' and root.parent!=root:root=root.parent
assert root.name=='agent-session-desk','Run archived verifier from repository'
e=root/'docs/testing/results/2026-11-24-long-cards-evidence';local=e/'local';artifacts=json.loads((e/'manifest.json').read_text())['artifacts']
for a in artifacts:assert hashlib.sha256((e/a['path']).read_bytes()).hexdigest()==a['sha256'],a['path']
fp=json.loads((local/'source-fingerprints.json').read_text());assert set(fp)=={str(p.relative_to(root)) for p in root.rglob('*.go')}|{'go.mod','go.sum'}
for name,sha in fp.items():assert hashlib.sha256((root/name).read_bytes()).hexdigest()==sha,name
checks=json.loads((local/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
assert 'no product or Windows tests were executed' in (local/'catalog-check.log').read_text()
status=json.loads((local/'windows-status.json').read_text());assert status['status']=='not_run' and not status['upload_attempted_this_slice'] and not status['native_test_started'] and not status['desktop_started']
src=json.loads((local/'upstream-source.json').read_text());assert src['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and not src['worktree'] and not src['production_source_modified'] and src['disabled_ui_tests']==456
source=json.loads((local/'upstream-frames.json').read_text());before=json.loads((local/'product-before.json').read_text());after=json.loads((local/'product-after.json').read_text());assert len(source)==len(before)==len(after)==8
comp=json.loads((local/'comparison.json').read_text());assert comp['summary']=={'title_before':1,'title_after':8,'metadata_before':6,'metadata_after':8,'all_widths_exact':8}
for s,b,a,c in zip(source,before,after,comp['cases']):
 assert s['name']==b['name']==a['name']==c['name']
 assert c['title_before']==(b['plain'][0]==s['crop_padded_for_product_surface'])
 assert c['title_after']==(a['plain'][0]==s['crop_padded_for_product_surface'])
 assert c['metadata_before']==(b['plain'][1:]==s['metadata_rows_padded'])
 assert c['metadata_after']==(a['plain'][1:]==s['metadata_rows_padded'])
 assert all(w==a['width'] for w in a['line_widths'])
assert sum(c['source_final_pad_cells'] for c in comp['cases'])==2
old=json.loads((local/'before-production.json').read_text())
for name,sha in old['overlaid_production_files'].items():assert hashlib.sha256((local/(name+'.before.txt')).read_bytes()).hexdigest()==sha
for name in ['card-red.log','inert-basename-red.log']:assert 'FAIL' in (local/name).read_text()
for name in ['card-green.log','inert-basename-green.log']:assert 'ok ' in (local/name).read_text()
assert 'PASS: TestASDLongCardReference' in (local/'upstream-reference-final.log').read_text()
hashes=json.loads((local/'binary-hashes.json').read_text());assert len(hashes)==2 and all(re.fullmatch('[0-9a-f]{64}',h) for h in hashes.values())
expected={a['path'] for a in artifacts}|{'manifest.json','verification.json','.gitattributes'};actual={str(p.relative_to(e)) for p in e.rglob('*') if p.is_file()};assert actual<=expected,actual-expected
result=dict(result='passed_local_evidence_consistency',artifacts=len(artifacts),source_fingerprints=len(fp),source_text_contract_cases=len(source),title_before=1,title_after=8,metadata_before=6,metadata_after=8,windows=status['status'],scope='normalized text crop/wrap, not raw full-frame/RGB, native, physical mouse/IME, agent, resource or release acceptance')
(e/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
