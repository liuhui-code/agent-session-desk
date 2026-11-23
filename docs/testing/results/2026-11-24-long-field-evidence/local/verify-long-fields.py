"""Check archived bounded evidence; does not execute product or Windows tests."""
from pathlib import Path
from collections import Counter
import hashlib,json,re,unicodedata
E=Path('docs/testing/results/2026-11-24-long-field-evidence')
def read(n):return json.loads((E/n).read_text())
def sha(b):return hashlib.sha256(b).hexdigest()
m=read('manifest.json');seen=set()
for item in m['artifacts']:
 p=item['path'];assert p not in seen and '..' not in Path(p).parts;seen.add(p);b=(E/p).read_bytes()
 assert sha(b)==item['archived_sha256'] and len(b)==item['bytes'],p
 assert Path(p).suffix not in ['.exe','.sqlite3','.db','.pid','.port','.zip','.key','.go'],p
 s=b.decode();assert not re.search(r'C:(?:\\)+'+'Us'+'ers'+r'(?:\\)+|/'+'Us'+'ers/',s,re.I),p
 if p.endswith('.json'):json.loads(s)
fp=read('local/source-fingerprints.json');assert len(fp)==210
for p,h in fp.items():assert sha(Path(p).read_bytes())==h,p
assert len([p for p in fp if p.startswith('docs/') and p.endswith('.go')])==2
assert len([p for p in fp if p.startswith('docs/') and p.endswith(('/go.mod','/go.sum'))])==4
old=read('local/before-caret/source-fingerprints.json');delta=read('local/caret-source-delta.json')
assert set(old)==set(fp) and {p for p in fp if fp[p]!=old[p]}==set(delta)=={'internal/ui/input_view.go','internal/ui/dialog_long_fields_test.go'}
assert sha((E/'local/input-before-caret.go.txt').read_bytes())==old['internal/ui/input_view.go']
checks=read('local/checks.json');assert len(checks)==6 and all(x['exit_code']==0 for x in checks)
assert 'all modules verified' in (E/'local/go-mod-verify.log').read_text()
assert 'all files <= 500 lines' in (E/'local/source-limits.log').read_text()
assert '"cases": 339' in (E/'local/catalog-check.log').read_text()
assert '"product_tests_executed_by_this_tool": 0' in (E/'local/catalog-check.log').read_text()
red=(E/'local/final-red.log').read_text();assert len(re.findall(r'^    --- FAIL:',red,re.M))==35 and len(re.findall(r'^    --- PASS:',red,re.M))==1
assert '--- FAIL: TestKeycapDisplayCaret' in (E/'local/caret-rune-red.log').read_text()
green=(E/'local/final-green.log').read_text();assert len(re.findall(r'^--- PASS:',green,re.M))==2 and len(re.findall(r'^    --- PASS:',green,re.M))==36
assert '--- FAIL:' not in green
hashes=read('local/binary-hashes.json');previous=read('local/before-caret/binary-hashes.json')
for folder,counts,h in [('native-ui',(175,403),hashes['long-field-ui-final.exe']),('native-ui-before-caret',(174,402),previous['long-field-ui.exe'])]:
 r=read(folder+'/outcome.json')['results'][0];assert (r['top_level_pass'],r['all_pass'])==counts and r['binary_sha256'].lower()==h
 assert r['exit_code']==r['fail']==r['skip']==r['remaining_exact_image']==0
 log=(E/folder/'ui-tests.log').read_text();assert len(re.findall(r'^--- PASS:',log,re.M))==counts[0] and len(re.findall(r'^\s*--- PASS:',log,re.M))==counts[1]
for n,captures,forced in [(1,0,1),(2,5,1),(3,31,1),(4,32,1),(5,32,0),(6,32,0)]:
 r=read(f'console-run{n}/outcome.json');assert r['captures']==captures and r['forced_product_ends']==forced and r['result']==('passed' if n==6 else 'failed')
 assert r['manager_ended'] and r['original_shell_ended'] and r['remaining_exact_product']==r['clipboard_api_calls']==0
 assert r['product_sha256'].lower()==(hashes['asd-long-field-final.exe'] if n>=5 else previous['asd-long-field.exe'])
assert len(list((E/'console-run1').glob('capture-*.json')))==1
for n,count in [(2,5),(3,31),(4,32),(5,32),(6,32)]:assert len(list((E/f'console-run{n}').glob('capture-*.json')))==count
r=read('console-run6/outcome.json');assert r['manager_exit_code']==0 and r['driver_ended'] and r['events']==70
assert read('console-run6/session-before.json')==read('console-run6/session-after.json')
assert read('console-run6/manager-identity.json')['handle_retained']
events=read('console-run6/events.json');assert len(events)==70 and Counter(x['kind'] for x in events)==Counter(text=13,key=27,resize=30)
keys=[x['details']['value'] for x in events if x['kind']=='key'];assert keys.count('Enter')==keys.count('q')==1
assert (E/'console-run6/shell.history.txt').read_text().splitlines()==["Write-Output 'LONG_DRAFT'"]
# Independent, deliberately fixture-specific column observer (not a Unicode engine).
def width(s):
 s=s.replace('👩‍💻','xx').replace('1️⃣','xx').replace('e\u0301','x')
 return sum(2 if unicodedata.east_asian_width(c) in 'WF' else 1 for c in s)
for i in range(1,33):
 v=read(f'console-run6/capture-{i:03d}.json');assert v['number']==i
 if i<=30:
  size=[(120,32),(63,25),(49,25),(120,12),(120,32)][(i-1)%5];assert (v['width'],v['height'])==size
  row=v['text'].splitlines()[v['cursor_y']];at=row.index('Q');assert width(row[:at+1])==v['cursor_x'],(i,row,v['cursor_x'])
 else:assert len(re.findall(r'(?m)│\s*LONG_DRAFT\s*│\s*$',v['text']))==1
observed=read('local/final-audit.json');assert observed['read_only'] and observed['original_alive']==observed['remaining_exact_images']==0
assert len(observed['originals'])==18 and all(not x['original_alive'] for x in observed['originals'])
assert len(observed['images'])==4 and all(x['remaining']==0 for x in observed['images'])
summary={'result':'passed_bounded_evidence_consistency','artifacts':len(seen),'source_fingerprints':len(fp),'dialog_cases':36,'model_size_states':252,'rune_boundary_cases':5,'final_native_ui':[175,403],'prior_native_ui':[174,402],'console_final':{'events':70,'captures':32,'forced_product_ends':0,'manager_exit_code':0,'clipboard_api_calls':0},'original_identities_ended':18,'earlier_console_failed_attempts':5,'scope':'not desktop, IME, full fidelity, real-agent, resource or release acceptance'}
(E/'verification.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n');print(json.dumps(summary,ensure_ascii=False))
