"""Verify archived evidence consistency, not product/desktop execution."""
from pathlib import Path
import hashlib,json,re,subprocess,sys
E=Path.cwd()/'docs/testing/results/2026-11-23-main-screen-evidence'
def read(n):return json.loads((E/n).read_text())
def sha(b):return hashlib.sha256(b).hexdigest()
def rows(v):
    if isinstance(v,dict) and set(v)=={'value','Count'}:assert v['Count']==len(v['value']);return v['value']
    assert isinstance(v,list);return v
m=read('manifest.json');seen=set()
for a in m['artifacts']:
    p=a['path'];assert p not in seen and '..' not in Path(p).parts;seen.add(p);b=(E/p).read_bytes();assert sha(b)==a['archived_sha256'] and len(b)==a['bytes']
    assert Path(p).suffix not in ('.exe','.sqlite3','.key','.port','.zip')
    if p.endswith('.png'):assert a['original_sha256']==a['archived_sha256'];continue
    assert not re.search('liu'+' xun yuan|/Users/'+'liuhui',b.decode(),re.I),p
fp=read('local/source-fingerprints.json');assert len(fp)==209
for p,h in fp.items():assert sha(Path(p).read_bytes())==h,p
assert len([p for p in fp if p.startswith('docs/') and p.endswith(('/go.mod','/go.sum'))])==4
assert len([p for p in fp if p.startswith('docs/') and p.endswith('.go')])==2
source=read('local/upstream-source.json');assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status']=='' and len(source['files'])==10
ref=read('local/upstream-main-reference.json');assert len(ref['frames'])==27 and len(ref['dividers'])==4
for n in ('product-main-reference.json','product-main-before.json'):
    v=read('local/'+n);assert len(v['cards'])==27 and len(v['frames'])==72
out=subprocess.run([sys.executable,str(E/'local/compare-main-roles.py'),str(E/'local')],capture_output=True,text=True,check=True)
r=read('local/role-comparison.json');assert (r['card_background_cells'],r['highlight_cells'],r['before_holes'],r['after_holes'])==(3591,2394,108,0)
assert r['known_accent_serialization_delta_modes']==r['before_divider_differences']==['terminal','hover','drag']
for mode in ('terminal','hover','drag'):
    d=next(x for x in r['divider_roles'] if x['mode']==mode);assert d['source_cell']['foreground']==[121,162,247] and d['product_cell']['foreground']==[122,162,247]
n=read('native-allmotion/outcome.json')['results'];assert len(n)==1;n=n[0]
assert (n['top_level_pass'],n['all_pass'])==(173,365) and n['exit_code']==n['fail']==n['skip']==n['remaining_exact_image']==0
assert n['binary_sha256'].lower()==read('local/binary-hashes.json')['main-allmotion-ui.exe']
log=(E/'native-allmotion/ui-tests.log').read_text();assert len(re.findall(r'^--- PASS:',log,re.M))==173 and len(re.findall(r'^\s*--- PASS:',log,re.M))==365
assert not re.search(r'^\s*--- (?:FAIL|SKIP):',log,re.M)
assert read('native/setup-outcome.json')['is_administrator'] is False
assert read('native/first-flow-failure.json')['sequence']==read('native/flow-failure.json')['sequence']==1
assert read('native/desktop/commands/001.json')=={'kind':'state'}
assert 'GetDataObject' in read('native/desktop/driver-error.json')['error']
assert read('native/desktop-run2/driver-error.json')['error']=='Foreground changed; input refused'
for name in ('desktop','desktop-run2'):
    for sub in ('responses','done'):assert not list((E/'native'/name/sub).glob('*.json'))
assert not list((E/'native').rglob('*.png')) and not (E/'native/flow-result.json').exists()
assert '[Windows.Forms.Clipboard]::' not in (E/'local/main-parity-desktop-driver.ps1').read_text()
env=read('native/desktop-run2/input-environment.json');assert env['clipboard_operations']==0 and env['clipboard_restoration_claimed'] is False and env['caps_before'] is env['caps_after'] is True
pre=read('native/driver-resume-preflight.json');assert pre['accepted_actions']==pre['replayed_mutations']==0 and pre['queued_readonly']==1
assert (E/'native/ui-v3-before.toml').read_bytes()==(E/'native/ui-final.toml').read_bytes()
a=read('native/final-audit.json');assert a['result']=='cleanup_passed_desktop_blocked' and a['accepted_ui_actions']==a['private_images_alive']==a['forced_count']==0
assert a['remaining_sessions']=='[]' and a['preferences_unchanged'] and a['run2_clipboard_operations']==0
assert a['first_clipboard_restoration'].startswith('unknown:') and a['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert sorted(a['tasks_absent'])==sorted('AsdMainParity'+n+'20261123' for n in ('Setup','Desktop','Read','Close'))
ids=rows(a['original_processes']);assert len(ids)==5 and all(x['original_ended'] for x in ids)
assert {x['kind'] for x in ids}=={'Copy_A','Copy_B','manager','terminal','outer-launcher'}
close=read('native/blocked-ui-close.json');assert close['close_window_requested'] and close['remaining_registered']==0 and not close['forced']
captures=rows(read('native/setup-captures.json'));assert len(captures)==2
originals=rows(read('native/sessions.json'));stops=rows(read('native/blocked-owned-stops.json'));assert {(x['session_id'],x['run_id']) for x in stops}=={(x['session_id'],x['run_id']) for x in originals}
for c in captures:assert c['numbered_ids']==list(range(1,41)) and 'MAIN_A:' not in c['text'] and 'MAIN_B:' not in c['text']
for name in ('go-test.log','go-race.log','go-vet.log','go-mod-verify.log'):assert not re.search(r'^FAIL(?:\s|$)',(E/'local'/name).read_text(),re.M)
assert 'all modules verified' in (E/'local/go-mod-verify.log').read_text()
assert 'all files <= 500 lines' in (E/'local/source-limits.log').read_text()
catalog=(E/'local/catalog-check.log').read_text();assert '"cases": 339' in catalog and '"product_tests_executed_by_this_tool": 0' in catalog
actual=read('local/desktop-final/actual-build-hashes.json')
assert actual['asd-main-allmotion.exe']==read('local/binary-hashes.json')['asd-main-allmotion.exe'] and actual['main-allmotion-ui.exe']==read('local/binary-hashes.json')['main-allmotion-ui.exe']
assert actual['asd-main-mouse-probe.exe']=='fd357e9c6537a1a6a32ba7f04a6ad06859fd0cc3eae884fee6bcbe30fa6685c9'
# Prior native build remains separate from the final AllMotion suite.
prior=read('native/native-tests/outcome.json')['results'][0]
assert (prior['top_level_pass'],prior['all_pass'])==(173,365) and prior['fail']==prior['skip']==prior['exit_code']==0
assert prior['binary_sha256'].lower()==read('local/binary-hashes-before-allmotion.json')['main-parity-ui.exe']
old=read('local/source-fingerprints-before-allmotion.json');delta=read('local/allmotion-source-delta.json')
assert set(old)==set(fp) and {p for p in fp if fp[p]!=old[p]}==set(delta)=={'internal/ui/home_view_frame.go','internal/ui/workspace_lifecycle_test.go','internal/ui/workspace_main_parity_test.go'}
for p,d in delta.items():assert d=={'before':old[p],'after':fp[p]}
assert all(f['mouse_mode']==2 for f in read('local/product-main-reference.json')['frames'])
# Final desktop is distinct from the earlier blocked root.
f=read('native-final/flow-result.json');assert f['result']=='passed' and f['commands']==269 and f['native_captures']==16
from collections import Counter
receipts={};kinds=Counter();shots={}
for folder,count in [('desktop',74),('desktop-continuation',75),('desktop-native',120)]:
    dd=E/'native-final'/folder
    rr=sorted((dd/'responses').glob('*.json'));done=sorted((dd/'done').glob('*.json'));assert len(rr)==len(done)==count
    for rp,dp in zip(rr,done):
        v=json.loads(rp.read_text());c=json.loads(dp.read_text());c.pop('original_json_sha256');assert v['command']==c and v['ok'];i=int(rp.stem);assert i not in receipts;receipts[i]=v;kinds[c['kind']]+=1
        if 'state' in v:assert 'text' not in v['state'] and 'tree' not in v['state'] and v['state']['terminal_text_redacted']
        if c['kind']=='capture':shots[i]=c['label']
assert sorted(receipts)==list(range(1,270))
assert kinds==Counter(state=187,text=16,keys=31,clickNavRow=2,divider=15,capture=12,resize=5,finish=1)
assert not receipts[164]['state']['divider_thick'] and receipts[167]['state']['divider_thick'] and receipts[169]['state']['divider_thick'] and not receipts[172]['state']['divider_thick']
assert [(receipts[i]['command']['width'],receipts[i]['command']['height']) for i in receipts if receipts[i]['command']['kind']=='resize']==[(1473,657),(600,539),(400,539),(1113,277),(1113,657)]
assert set(p.name for p in (E/'native-final').rglob('*.png'))=={'main-nested-selected-120-public.png','main-divider-hover-120-public.png','main-divider-pressed-120-public.png'}
for i in (164,167,169):
    label=shots[i];b=(E/'native-final/desktop-native'/(label+'.png')).read_bytes();assert b.startswith(b'\x89PNG\r\n\x1a\n')
    import struct
    assert list(struct.unpack('>II',b[16:24]))==receipts[i]['state']['rect'][2:]
left={v['session_id']:v for v in rows(f['before'])};right={v['session_id']:v for v in rows(f['after'])};assert left.keys()==right.keys() and len(left)==2
for sid in left:
    for key in ('run_id','backend_id','title','command','terminal_state'):assert left[sid][key]==right[sid][key]
assert rows(f['before_processes'])==rows(f['after_processes']);procs={p['label']:p for p in rows(f['before_processes'])}
o=rows(read('native-final/native-observations.json'));assert len(o)==16
for v in o:
    assert v['complete'] and v['numbered_ids']==list(range(1,41)) and v['terminal_text_redacted'] and 'capture_text' not in v and 'screen_text' not in v
    text='\n'.join(v['fixture_only_lines']);assert [int(n) for n in re.findall(r'^history line (\d+)$',text,re.M)]==list(range(1,41))
    letter=v['label'][0];owner=procs['Copy '+letter];sid=next(s for s in left if left[s]['title']=='Copy '+letter)
    assert v['session_id']==sid and v['run_id']==left[sid]['run_id']
    result=re.findall(r'^MAIN_'+letter+r':(\d+)\s*$',text,re.M)
    if v['label']==letter+'-explicit-enter':assert result==[str(owner['pid'])]
    else:assert not result and any(z.startswith("> Write-Output ('MAIN_"+letter+":'+$PID)") for z in v['fixture_only_lines'])
assert f['preferences_before']==f['preferences_after']
assert (E/'native-final/ui-v3-before.toml').read_bytes()==(E/'native-final/ui-final.toml').read_bytes()
a=read('native-final/final-audit.json');assert a['result']=='passed_bounded_cleanup' and a['private_images_alive']==a['remaining_synthetic_launchers']==a['forced_product_ends']==0 and a['controller_replacements']==1 and a['remaining_sessions']=='[]'
assert len(rows(a['original_processes']))==7 and all(v['original_ended'] for v in rows(a['original_processes']))
assert Counter(v['kind'] for v in rows(a['original_processes']))==Counter({'Copy_A':1,'Copy_B':1,'formal-ui':3,'mouse-diagnostic-ui':1,'terminal':1})
assert len(a['tasks_absent'])==5 and a['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert a['input_environment']['clipboard_operations']==0 and a['input_environment']['caps_before'] is a['input_environment']['caps_after'] is False
assert read('native-final/flow-failure.json')['sequence']==74 and read('native-final/resume-failure.json')['sequence']==125 and read('native-final/native-resume-failure.json')['sequence']==184 and read('native-final/responsive-resume-failure.json')['sequence']==237
assert read('native-final/flow-resume.json')['previous_sequence']==74 and read('native-final/responsive-resume.json')['previous_sequence']==184 and read('native-final/final-tail-resume.json')['previous_sequence']==237
assert all(read('native-final/'+n)['replayed_mutations']==0 for n in ('flow-resume.json','responsive-resume.json','final-tail-resume.json'))
move=read('native-final/native-move-probe3.json');assert move['result']=='observed' and [v['sent'] for v in move['points']]==[1,1]
assert 'Motion x=43 y=2 button=0' in move['trace'] and 'Motion x=0 y=2 button=0' in move['trace']
assert read('native-final/native-move-probe2.json')['error']=='Foreground not owned; no input'
result={'result':'passed_bounded_evidence_consistency','artifacts':len(seen),'go_module_fingerprints':209,'historical_reference_mains':2,'historical_nested_module_files':4,'source_card_pairs':27,'background_cells':3591,'old_holes':108,'new_holes':0,'divider_roles':4,'known_serialization_differences':3,'product_main_frames':72,'final_native_ui':[173,365],'earlier_desktop':'blocked, zero accepted UI actions','final_desktop':{'protocol_receipts':269,'actions':dict(kinds),'native_captures':16,'public_png':3,'registered_originals_ended':7,'forced_product_ends':0,'controller_replacements':1,'replayed_mutations':0},'first_clipboard_restoration':'unknown','final_clipboard_api_calls':0,'privacy':'Script-external terminal body excluded, no author or no-interference inference; embedded source JSON hashes refer to decoded UTF-8 text','scope':'Bounded source/model/native-test and formal desktop evidence, not complete fidelity, real-agent, resource or release acceptance'}
(E/'verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
