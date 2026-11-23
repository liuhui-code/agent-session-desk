from pathlib import Path
import hashlib,json,re,collections
root=Path('docs/testing/results/2026-11-23-groups-desktop-evidence')
def read(path):return json.loads((root/path).read_text())
x=read('manifest.json');expected={f['path']for f in x['files']};actual={p.relative_to(root).as_posix()for p in root.rglob('*')if p.is_file()}-{'manifest.json','README.md'}
assert expected==actual,{'missing':sorted(expected-actual),'extra':sorted(actual-expected)}
for f in x['files']:
 data=(root/f['path']).read_bytes();assert hashlib.sha256(data).hexdigest()==f['sha256'],f['path']
 if f['binary_unchanged']:assert f['source_sha256']==f['sha256']
 else:
  text=data.decode('utf-8');assert str(Path.home()) not in text,f['path']
  assert not re.search(r'(?i)[a-z]:\\+users\\+(?!<WINDOWS_TEST_USER>)[^\\\s"]+',text),f['path']
source=read('local/product-source-after.json');assert len(source['files'])==174
for f in source['files']:assert hashlib.sha256(Path(f['path']).read_bytes()).hexdigest()==f['sha256'],f['path']
flow=read('native/flow-result.json');assert flow['result']=='passed'and flow['commands']==174
kinds=collections.Counter()
for i in range(1,175):
 name=f'{i:03}.json';r=read('native/desktop/responses/'+name);done=read('native/desktop/done/'+name)
 assert r['ok']and r['command']==done,(name,r)
 kinds[r['command']['kind']]+=1
assert kinds['state']==60 and kinds['capture']==16
before={v['session_id']:v for v in flow['before']};after={v['session_id']:v for v in flow['after']}
assert len(before)==2 and len(after)==3
for sid,v in before.items():
 for key in ['session_id','run_id','backend_id','command']:assert after[sid][key]==v[key],(sid,key)
procs={v['label']:v for v in flow['after_processes']};assert len(procs)==3
for v in flow['before_processes']:
 for key in ['pid','filetime','path','cwd']:assert v[key]==procs[v['label']][key],key
assert len({v['pid']for v in procs.values()})==3
assert procs['Group C']['cwd']==flow['child_group']['DefaultDir']
observations=read('native/native-observations.json');assert len(observations)==18
panes={}
for o in observations:
 assert o['complete']and o['numbered_ids']==list(range(1,41)),o['label']
 sid=o['session_id'];assert o['run_id']==after[sid]['run_id']
 if sid in panes:assert panes[sid]==o['pane']
 panes[sid]=o['pane']
 title=after[sid]['title'];letter={'Copy A':'A','Copy B':'B','Group C':'C'}[title];marker=f'GROUP_{letter}:{procs[title]["pid"]}'
 if o['label'].endswith('explicit-enter'):assert re.search(r'^'+re.escape(marker)+r'\s*$',o['capture_text'],re.M)
 else:
  assert not re.search(r'^GROUP_'+letter+r':\d+\s*$',o['capture_text'],re.M),o['label']
  lines=[s.strip()for s in o['capture_text'].splitlines()if s.strip()]
  assert lines[-1]==f"> Write-Output ('GROUP_{letter}:'+$PID)",(o['label'],lines[-1])
metadata=read('native/metadata-observations.json');assert len(metadata)==8
runs={}
for item in metadata:
 v=item['value'];assert v['Version']==4
 assert len({g['Scope']for g in v['Groups']})==1
 for run in v['Runs']:
  key=run['ID'];stable=(run['SessionID'],run['BackendID'],run['Command'],run['RecordedNS'])
  if key in runs:assert runs[key]==stable,key
  runs[key]=stable
final=next(m['value']for m in metadata if m['label']=='reopened');backup=next(m['value']for m in metadata if m['label']=='backup-groups')
assert len(final['Groups'])==len(backup['Groups'])==3 and len(final['Members'])==len(backup['Members'])==3 and len(final['Runs'])==3
assert {g['ID']for g in final['Groups']}=={g['ID']for g in backup['Groups']}
assert flow['child_group']['ParentID']==flow['root_group']['ID']
member={m['SessionID']:m for m in final['Members']}
for sid,v in after.items():
 target=flow['root_group']['ID']if v['title']=='Copy B'else flow['child_group']['ID']
 assert member[sid]['GroupID']==target and member[sid]['Revision']==1
prefs=read('native/preferences.json');assert len(prefs)==8
fold=next(v for v in prefs if v['label']=='reopen-fold');assert fold['folds']==flow['root_group']['ID']and fold['density']=='full'
assert len(list((root/'native/desktop').glob('*.png')))==15
assert not (root/'native/desktop/new-in-child-default-directory-public.png').exists()
audit=read('native/final-audit.json');assert audit['result']=='passed'and audit['private_images_alive']==0 and audit['remaining_sessions']=='[]'
assert len(audit['original_processes'])==6 and all(v['original_ended']for v in audit['original_processes'])
for v in audit['original_processes']:assert isinstance(v['pid'],int)and isinstance(v['filetime'],int)and v['filetime']>2**53 and isinstance(v['path'],str)
assert len(audit['tasks_absent'])==2
assert audit['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
clip=read('native/desktop/clipboard-restored.json');assert clip['restored']and clip['text_restored']and clip['format_names_restored']and clip['caps_after']==clip['caps_before']
ui=read('local/native-ui-result.json');assert ui['exit_code']==ui['fail']==ui['skip']==ui['remaining_exact_image']==0
text=(root/'local/native-ui-tests.log').read_text();assert len(re.findall(r'^--- PASS:',text,re.M))==ui['top_level_pass']==131
assert len(re.findall(r'^\s*--- PASS:',text,re.M))==ui['all_pass']==305
assert all(v['exit_code']==0 for v in read('local/check-exits.json'))
raw=read('local/stock-placeholder-repro.json');assert raw['nul_count']==4 and raw['value']=='' and raw['cursor_x']==2
frames=read('local/product-dialog-frames.json');assert len(frames)==15 and all('\0'not in f['content']for f in frames)
geometry=read('local/dialog-geometry.json');a,b=geometry[:15],geometry[15:];assert len(a)==len(b)==15
matches=0
for s,p in zip(a,b):
 assert (s['columns'],s['mode'],s['field'])==(p['columns'],p['mode'],p['field'])
 assert s['box'][0]==p['box'][0]and s['box'][2]==p['box'][2]
 if s['box']==p['box']:matches+=1
 else:assert s['columns']==49 and s['mode']in['root','subgroup']and s['box'][3]==p['box'][3]+1
assert matches==11
print(json.dumps({'artifacts':len(expected),'receipts':174,'native_complete_captures':18,'metadata_snapshots':8,'preference_snapshots':8,'unchanged_png':15,'originals_ended':6,'native_ui_top':131,'native_ui_all':305,'source_product_pairs':15,'source_files':174,'result':'verified'}))
