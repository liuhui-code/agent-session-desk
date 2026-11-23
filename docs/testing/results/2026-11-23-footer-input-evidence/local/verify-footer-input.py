"""Verify archived evidence consistency, not product execution."""
from collections import Counter
import hashlib,json,re,struct
from pathlib import Path
REPO=Path.cwd();E=REPO/'docs/testing/results/2026-11-23-footer-input-evidence'
def read(p):return json.loads((E/p).read_text())
def sha(b):return hashlib.sha256(b).hexdigest()
def rows(v):
    if isinstance(v,dict) and set(v)=={'value','Count'}:assert v['Count']==len(v['value']);return v['value']
    assert isinstance(v,list);return v
m=read('manifest.json');paths=set()
for a in m['artifacts']:
    p=a['path'];assert p not in paths and '..' not in Path(p).parts;paths.add(p);b=(E/p).read_bytes()
    assert sha(b)==a['archived_sha256'] and len(b)==a['bytes'];assert Path(p).suffix not in ('.exe','.sqlite3','.key','.port')
    if p.endswith('.png'):assert a['original_sha256']==a['archived_sha256']
    else:assert not re.search('liu'+' xun yuan|/Users/'+'liuhui',b.decode('utf-8'),re.I),p
fp=read('local/source-fingerprints.json');assert len(fp)==204
for p,h in fp.items():assert sha((REPO/p).read_bytes())==h,p
old=read('local/source-fingerprints-before-target.json');assert set(fp)==set(old)
assert {p for p in fp if fp[p]!=old[p]}=={'internal/ui/workspace_footer.go','internal/ui/workspace_footer_test.go'}
source=read('local/upstream-source.json');assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status']=='' and len(source['files'])==19
ref=read('local/upstream-reference.json');product=read('local/product-reference.json')
assert ref['default_footer']=='full' and len(ref['frames'])==len(product['frames'])==24
assert len(ref['paste_cases'])==10 and all(not v['visible'] and v['offset']==0 and v['all_after']==v['all'] for v in ref['paste_cases'])
role=read('local/role-comparison.json');assert role['result']=='passed' and role['known_accent_serialization_cells']==12
assert role['source_emitted_accent']==[121,162,247] and role['product_emitted_accent']==[122,162,247]
assert role['role_cells']=={'role_full':12,'role_compact':10,'role_badge':17}
assert read('local/color-serialization.json')['authored_hex']=='#7aa2f7'
for f in product['frames']:
    if f['mode']=='terminal' and f['columns']==49:assert '→' not in f['footer'] and ']' not in f['footer']
def native(prefix,hashfile,binary,top,total):
    result=read(prefix+'/native-tests/outcome.json')['results'];assert len(result)==1;r=result[0]
    assert (r['top_level_pass'],r['all_pass'])==(top,total)
    assert r['exit_code']==r['fail']==r['skip']==r['remaining_exact_image']==0
    assert r['binary_sha256'].lower()==read('local/'+hashfile)[binary]
    text=(E/prefix/'native-tests/ui-tests.log').read_text()
    assert len(re.findall(r'^--- PASS:',text,re.M))==top and len(re.findall(r'^\s*--- PASS:',text,re.M))==total
    assert not re.search(r'^\s*--- (?:FAIL|SKIP):',text,re.M)
native('native-full-flow','binary-hashes-before-target.json','footer-input-ui.exe',170,354)
native('native','binary-hashes.json','footer-label-ui.exe',171,355)
def flow(prefix,marker,count,task):
    f=read(prefix+'/flow-result.json');assert f['result']=='passed'
    rr=sorted((E/prefix/'desktop/responses').glob('*.json'));dd=sorted((E/prefix/'desktop/done').glob('*.json'))
    assert [p.stem for p in rr]==[f'{i:03}' for i in range(1,f['commands']+1)] and [p.name for p in rr]==[p.name for p in dd]
    kinds=Counter();commands=[];shots=[]
    for rp,dp in zip(rr,dd):
        r=json.loads(rp.read_text());c=json.loads(dp.read_text());assert r['ok'] and r['command']==c;commands.append(c);kinds[c['kind']]+=1
        if c['kind']=='capture':
            b=(E/prefix/'desktop'/(c['label']+'.png')).read_bytes();assert b.startswith(b'\x89PNG\r\n\x1a\n') and list(struct.unpack('>II',b[16:24]))==r['state']['rect'][2:];shots.append(c['label'])
    before={v['session_id']:v for v in rows(f['before'])};after={v['session_id']:v for v in rows(f['after'])};assert len(before)==2 and before==after
    processes=rows(f['before_processes']);assert len(processes)==2 and processes==rows(f['after_processes'])
    oo=rows(read(prefix+'/native-observations.json'));assert len(oo)==f['native_captures']==count
    for o in oo:
        assert o['complete'] and o['numbered_ids']==list(range(1,41));text=o['capture_text'];assert [int(v) for v in re.findall(r'^history line (\d+)\s*$',text,re.M)]==list(range(1,41))
        letter=o['label'][0];owner=processes[0 if letter=='A' else 1];sid=next(k for k,v in before.items() if v['title']=='Copy '+letter)
        assert o['session_id']==sid and o['run_id']==before[sid]['run_id']
        out=re.findall(r'^'+marker+'_'+letter+r':(\d+)\s*$',text,re.M)
        if o['label']==letter+'-explicit-enter':assert out==[str(owner['pid'])]
        else:
            assert not out
            suffix=(' # nrMdx?+' if marker=='FOOT' else ' # narrow') if letter=='A' else ''
            expected="> Write-Output ('"+marker+'_'+letter+":'+$PID)"+suffix
            assert [line.strip() for line in text.splitlines() if line.strip()][-1]==expected
    assert f['preferences_before'].lower()==f['preferences_after'].lower()
    assert (E/prefix/'ui-v3-before.toml').read_bytes()==(E/prefix/'ui-final.toml').read_bytes()
    audit=read(prefix+'/final-audit.json');assert audit['result']=='passed' and audit['private_images_alive']==0 and audit['remaining_sessions']=='[]'
    assert len(audit['stops'])==2 and {(v['session_id'],v['run_id']) for v in audit['stops']}=={(v['session_id'],v['run_id']) for v in before.values()}
    identities=rows(audit['original_processes']);assert len(identities)==5 and all(v['original_ended'] for v in identities)
    assert audit['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
    assert sorted(audit['tasks_absent'])==[task+'Desktop20261123',task+'Setup20261123']
    clip=audit['clipboard'];assert clip['restored'] and clip['text_restored'] and clip['format_names_restored'] and clip['caps_before'] is clip['caps_after'] is True and clip['nontext_content_verified'] is False
    return f,kinds,commands,shots
full,kinds,commands,shots=flow('native-full-flow','FOOT',12,'AsdFooterInput')
assert full['commands']==193 and full['help_paste_cases']==10 and full['preference_snapshots']==13 and full['explicit_fold_toggles']==2
assert kinds==Counter(state=112,capture=17,text=19,keys=24,resize=7,paste=13,finish=1) and len(shots)==17
prefs=rows(read('native-full-flow/preference-observations.json'));assert len(prefs)==13 and [v['label'] for v in prefs[2:12]]==[f'paste-{i}' for i in range(1,11)]
assert all(v['sha256']==full['preferences_before'] for v in prefs)
resume=read('native-full-flow/flow-resume.json');assert resume['previous_completed_sequence']==154 and resume['fresh_state_sequence']==155 and resume['replayed_mutations']==0
inspection=read('local/inspect-after-wait-failure.json');assert inspection['last_response']=='154.json' and inspection['task_state']=='Running' and len(inspection['driver_processes'])==1
assert inspection['driver_processes'][0]['pid']==resume['original_driver_pid'] and inspection['driver_processes'][0]['filetime']==resume['original_driver_filetime']
shape=read('local/resume-shape.json');assert shape['last_response']=='154.json' and shape['driver_live'] and shape['observed_count']==4 and shape['wrapped_count']==1 and shape['preferences_count']==12
assert read('native-full-flow/flow-failure.json')['sequence']==154
assert commands[154]['kind']=='state' and commands[155]['label']=='footer-one-tab-after-accepted-close-120-public' and commands[156]=={'kind':'text','text':'x'}
# Ten modal payloads, including literal tab and multiline, are consumed as one paste each.
pastes=[c['text'] for c in commands if c['kind']=='paste'];assert pastes[2:12]==['n','q','j','\t','line one\nline two']*2
final,fk,fc,fs=flow('native','LABEL',9,'AsdFooterLabel');assert len(fs)==5
for rp in (E/'native/desktop/responses').glob('*.json'):
    r=json.loads(rp.read_text());label=r['command'].get('label','')
    if label.endswith('-49-public'):
        lines=[v for v in r['state']['text'].splitlines() if '终端模式' in v];assert len(lines)==1 and '→' not in lines[0] and ']' not in lines[0] and 'F6' in lines[0] and 'Ctrl+PgUp/PgDn' in lines[0]
result={'result':'passed','artifacts':len(paths),'source_fingerprints':len(fp),'source_product_frame_pairs':24,'source_paste_cases':10,'final_native_ui':[171,355],'known_rgb_serialization_cells':12,'full_flow':{'commands':193,'actions':dict(kinds),'captures':12,'screenshots':17,'replayed_mutations':0},'final_label_flow':{'commands':final['commands'],'actions':dict(fk),'captures':9,'screenshots':5},'original_processes_ended_per_flow':5,'scope':'Bounded evidence consistency only; final label and earlier full-input build are distinct, no full UI/release acceptance'}
(E/'verification.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps(result,ensure_ascii=False))
