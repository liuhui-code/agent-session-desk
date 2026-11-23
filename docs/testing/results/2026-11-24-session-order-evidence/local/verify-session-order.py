"""Verify bounded archived evidence only; no product or Windows execution."""
from pathlib import Path
from collections import Counter
import hashlib,json,re
E=Path('docs/testing/results/2026-11-24-session-order-evidence')
def read(n):return json.loads((E/n).read_text())
def sha(b):return hashlib.sha256(b).hexdigest()
m=read('manifest.json');seen=set()
for a in m['artifacts']:
 p=a['path'];assert p not in seen and '..' not in Path(p).parts;seen.add(p);b=(E/p).read_bytes();assert sha(b)==a['archived_sha256'] and len(b)==a['bytes'],p
 assert Path(p).suffix not in {'.exe','.sqlite3','.db','.pid','.port','.key','.go','.zip'},p
 s=b.decode();assert not re.search(r'C:(?:\\)+'+'Us'+'ers'+r'(?:\\)+|/'+'Us'+'ers/',s,re.I),p
 if p.endswith('.json'):json.loads(s)
fp=read('local/source-fingerprints.json');before=read('local/source-fingerprints-before-membership.json');assert len(fp)==222 and set(fp)==set(before)
for p,h in fp.items():assert sha(Path(p).read_bytes())==h,p
assert len([p for p in fp if p.startswith('docs/') and p.endswith('.go')])==2
assert len([p for p in fp if p.startswith('docs/') and p.endswith(('/go.mod','/go.sum'))])==4
delta=read('local/membership-source-delta.json');assert {p for p in fp if fp[p]!=before[p]}==set(delta)=={'internal/ui/group_dialog.go','internal/ui/session_order_lifecycle_test.go'}
assert sha((E/'local/group-dialog-before-membership.go.txt').read_bytes())==before['internal/ui/group_dialog.go']
checks=read('local/checks.json');assert len(checks)==6 and all(x['exit_code']==0 for x in checks)
assert 'all modules verified' in (E/'local/go-mod-verify.log').read_text()
assert 'all files <= 500 lines' in (E/'local/source-limits.log').read_text()
assert '"cases": 339' in (E/'local/catalog-check.log').read_text()
assert '"product_tests_executed_by_this_tool": 0' in (E/'local/catalog-check.log').read_text()
for name in ['store-red','ui-red','position-validation-red','stamp-red','membership-red']:assert '--- FAIL:' in (E/'local'/f'{name}.log').read_text()
for name in ['store-green','stamp-green','membership-green','bindings']:assert '--- FAIL:' not in (E/'local'/f'{name}.log').read_text()
source=read('local/upstream-source.json');assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['license']=='MIT' and source['overlay_test_files']==840
states=read('local/upstream-order-states.json');assert [v['state'] for v in states]==['initial','up-boundary','swap-up','rebuild','swap-down','down-boundary']
assert [v['ids'] for v in states]==[['a','b'],['a','b'],['b','a'],['b','a'],['a','b'],['a','b']]
assert 'module lookup disabled by GOPROXY=off' in (E/'local/upstream-order.log').read_text()
assert '--- PASS: TestASDSessionOrderReference' in (E/'local/upstream-reference.log').read_text()
hashes=read('local/binary-hashes.json');oldhash=read('local/binary-hashes-before-membership.json');counts={}
for r in read('native/outcome.json')['results']:
 package=r['package'];want={'session':(9,9),'store':(73,115),'manager':(17,19),'ui':(181,427)}[package]
 assert (r['top_level_pass'],r['all_pass'])==want and r['exit_code']==r['fail']==r['skip']==r['remaining_exact_image']==0
 assert r['binary_sha256'].lower()==oldhash['session-order-'+package+'.exe']
 if package!='ui':assert hashes['session-order-'+package+'.exe']==r['binary_sha256'].lower();counts[package]=want
 log=(E/'native'/f'{package}.log').read_text();assert len(re.findall(r'^--- PASS:',log,re.M))==want[0] and len(re.findall(r'^\s*--- PASS:',log,re.M))==want[1]
r=read('native-final/outcome.json')['results'][0];assert (r['top_level_pass'],r['all_pass'])==(182,428) and r['binary_sha256'].lower()==hashes['session-order-ui.exe'] and r['exit_code']==r['fail']==r['skip']==r['remaining_exact_image']==0
log=(E/'native-final/ui.log').read_text();assert len(re.findall(r'^--- PASS:',log,re.M))==182 and len(re.findall(r'^\s*--- PASS:',log,re.M))==428;counts['ui']=(182,428)
for n,passed,captures,events,forced,registered in [(1,False,5,10,1,5),(2,False,8,17,1,6),(3,True,9,20,0,6),(4,False,0,1,0,1),(5,True,12,32,0,6)]:
 r=read(f'console-run{n}/outcome.json');assert r['result']==('passed'if passed else'failed') and (r['captures'],r['events'],r['forced_product_ends'])==(captures,events,forced)
 assert len(r['originals'])==registered and all(x['original_ended']for x in r['originals'])
 assert r['clipboard_api_calls']==r['remaining_exact_product']==0
 assert r['product_sha256'].lower()==(oldhash['asd-session-order.exe']if n<=3 else hashes['asd-session-order.exe'])
 assert len(list((E/f'console-run{n}').glob('capture-*.json')))==captures
for n in [1,2,3,5]:
 assert read(f'console-run{n}/metadata-before.json')['schema_version']==4 and read(f'console-run{n}/metadata-after.json')['schema_version']==5
 reject=read(f'console-run{n}/legacy-v5-rejection.json');assert reject['exit_code']!=0 and reject['sha256_before']==reject['sha256_after']
 assert read(f'console-run{n}/restored-order.json')==read(f'console-run{n}/order-final.json')
folder='console-run5/';events=read(folder+'events.json');assert Counter(x['kind']for x in events)==Counter(text=11,key=19,**{'manager-exited':2})
assert all(x['details']['exit_code']==0 for x in events if x['kind']=='manager-exited')
original=read(folder+'sessions-before.json');after=read(folder+'sessions-after.json');a,b=[x['session_id']for x in original]
for first,last in zip(sorted(original,key=lambda x:x['session_id']),sorted(after,key=lambda x:x['session_id'])):
 for key in ['session_id','run_id','backend_id','title','command']:assert first[key]==last[key]
for name,want in [('order-up',[b,a]),('order-down',[a,b]),('order-before-membership',[b,a]),('moved-back-order',[a,b]),('order-final',[b,a]),('restored-order',[b,a])]:assert [v['session_id']for v in read(folder+name+'.json')]==want
moved=next(x for x in read(folder+'moved-to-other.json')if x['session_id']==b);assert moved['group_id'] and moved['group_revision']==1 and moved.get('order_revision',0)==0
back=next(x for x in read(folder+'moved-back-order.json')if x['session_id']==b);assert not back.get('group_id') and back['group_revision']==2 and back.get('order_revision',0)==0
final=read(folder+'order-final.json');assert [(x.get('order',0),x['order_revision'])for x in final]==[(0,1),(1,4)]
for label in ['A','B']:assert (E/folder/(label+'.history.txt')).read_text().splitlines()==["Write-Output ('ORDER_"+label+":'+$PID)"]
identities=read(folder+'identities.json');assert len(identities)==6 and all(x['handle_retained']for x in identities)
for label,i in [('A',11),('B',12)]:
 pid=next(x['pid']for x in identities if x['kind']=='shell-'+label);capture=read(folder+f'capture-{i:03d}.json');assert len(re.findall(r'(?m)│\s*ORDER_'+label+':'+str(pid)+r'\s*│\s*$',capture['text']))==1
for i in range(1,13):assert read(folder+f'capture-{i:03d}.json')['number']==i
h1={r['session']['session_id']:r for r in read(folder+'history-before.json')};h2={r['session']['session_id']:r for r in read(folder+'history-after.json')};assert set(h1)==set(h2)=={a,b}
for sid in h1:
 assert h1[sid]['recorded_at']==h2[sid]['recorded_at'] and h1[sid].get('stop_confirmed_at')==h2[sid].get('stop_confirmed_at')
 for key in ['session_id','run_id','backend_id','title','command']:assert h1[sid]['session'][key]==h2[sid]['session'][key]
audit=read('local/final-audit.json');assert audit['read_only'] and audit['registered_remaining']==audit['registry_remaining']==audit['remaining_exact_images']==0
assert len(audit['registered'])==24 and all(x['original_ended']for x in audit['registered'])
assert all(x['original_ended']for x in audit['registry']) and len(audit['images'])==5 and all(x['remaining']==0 for x in audit['images'])
assert len(audit['shared_candidate_images'])==2 and all(x['remaining']==0 for x in audit['shared_candidate_images'])
summary={'result':'passed_bounded_evidence_consistency','artifacts':len(seen),'source_fingerprints':222,'upstream_reference_states':6,'native_final':counts,'prior_ui':[181,427],'final_console':{'events':32,'captures':12,'forced_product_ends':0,'manager_exits_zero':2,'clipboard_api_calls':0},'registered_originals_ended':24,'earlier_incomplete_registrations_preserved':True,'scope':'not full upstream, desktop, physical modifiers, IME, real-agent, resource or release acceptance'}
(E/'verification.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n');print(json.dumps(summary,ensure_ascii=False))
