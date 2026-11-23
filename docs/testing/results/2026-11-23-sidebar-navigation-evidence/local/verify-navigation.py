from pathlib import Path
import hashlib,json,re

evidence=Path('docs/testing/results/2026-11-23-sidebar-navigation-evidence')
manifest=json.loads((evidence/'manifest.json').read_text())
for item in manifest['files']:
 data=(evidence/item['path']).read_bytes()
 assert len(data)==item['archived_bytes'],item['path']
 assert hashlib.sha256(data).hexdigest()==item['archived_sha256'],item['path']
 assert b'<LOCAL_TEST_HOME>' not in data and b'<WINDOWS_TEST_USER>' not in data,item['path']
source=json.loads((evidence/'local/product-source.json').read_text())
for item in source['files']:
 assert hashlib.sha256(Path(item['path']).read_bytes()).hexdigest()==item['sha256'],item['path']
results=json.loads((evidence/'native/results.json').read_text())
assert results['windows']=='10.0.19045.0'
assert results['live_exact_paths']==[]
hashes=json.loads((evidence/'local/binary-hashes.json').read_text())
counts=[]
for r in results['tests']:
 text=(evidence/'native'/(r['name']+'.log')).read_text()
 assert r['sha256']==hashes[r['name']]
 assert r['exit_code']==r['fail']==r['skip']==0
 assert len(re.findall(r'^--- PASS:',text,re.M))==r['top_level_pass']
 assert len(re.findall(r'^\s*--- PASS:',text,re.M))==r['all_pass']
 assert not re.search(r'^\s*--- (FAIL|SKIP):',text,re.M)
 counts.append([r['name'],r['top_level_pass'],r['all_pass']])
assert counts==[['sidebar-ui-tests.exe',120,291],['sidebar-config-tests.exe',15,15]]
frames=json.loads((evidence/'local/upstream-header-frames.json').read_text())
assert len(frames)==32
for name in ['red-config.log','red-grouped.log','red-mouse.log','red-quit.log']:
 assert 'FAIL' in (evidence/'local'/name).read_text(),name
assert 'FAIL' not in (evidence/'local/quit-first-pass-input-masked.log').read_text()
for name in ['green-final.log','go-test.log','go-race.log']:
 text=(evidence/'local'/name).read_text()
 assert not re.search(r'^(FAIL|--- FAIL:)',text,re.M),name
 assert 'internal/ui' in text and 'internal/config' in text,name
assert (evidence/'local/go-vet.log').read_bytes()==b''
assert 'all modules verified' in (evidence/'local/go-mod-verify.log').read_text()
catalog=(evidence/'local/catalog-check.log').read_text().splitlines()[0]
assert json.loads(catalog)['cases']==339
summary={'verified_archived_files':len(manifest['files']),'current_source_fingerprints':len(source['files']),'upstream_component_frames':len(frames),'native_counts':counts,'native_desktop_executed':False,'catalog_cases':339,'scope':'Partial default-group navigation; custom groups and full release gates remain open.'}
(evidence/'checks-summary.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(summary,ensure_ascii=False))
