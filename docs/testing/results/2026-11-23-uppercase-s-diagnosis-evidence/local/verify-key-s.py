from pathlib import Path
from collections import Counter,defaultdict
import json,hashlib,re,struct
root=Path('docs/testing/results/2026-11-23-uppercase-s-diagnosis-evidence')
def normal(x):
 if isinstance(x,dict):
  if set(x)=={'value','Count'}:assert isinstance(x['value'],list) and len(x['value'])==x['Count'];return normal(x['value'])
  return {k:normal(v) for k,v in x.items()}
 if isinstance(x,list):return [normal(v) for v in x]
 return x
def read(path):return normal(json.loads((root/path).read_text()))
def identity(x):
 assert isinstance(x['pid'],int) and isinstance(x['filetime'],int) and isinstance(x['path'],str)
 return x['pid'],x['filetime'],x['path'].lower()
manifest=read('manifest.json')
for f in manifest['files']:
 data=(root/f['path']).read_bytes();assert hashlib.sha256(data).hexdigest()==f['archived_sha256'] and len(data)==f['archived_bytes'],f['path']
 if not f['path'].endswith('.png'):assert b'<LOCAL_TEST_HOME>' not in data and b'<WINDOWS_TEST_USER>' not in data,f['path']
source=read('local/source-before.json');assert source==read('local/source-after.json')
for entry in read('local/instrumentation-sources.json')['files']:
 path=Path(entry['path'].replace('<REPO>',str(Path.cwd())))
 assert hashlib.sha256(path.read_bytes()).hexdigest()==entry['sha256'],entry['path']
assert json.loads((root/'local/catalog-check.log').read_text().splitlines()[0])['cases']==339
assert 'all files <= 500 lines' in (root/'local/source-limits.log').read_text()
for f in source['files']:assert hashlib.sha256(Path(f['path']).read_bytes()).hexdigest()==f['sha256'],f['path']
assert source['original_product_sha256']=='f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277'
base=read('native/baseline-result.json');probe=read('native/probe-result.json');result=read('native/key-s-result.json');audit=read('native/final-audit.json')
assert all(not x['sendkeys_opened_settings'] and x['unicode_opened_settings'] for x in base['trials']) and len(base['trials'])==3
assert [(x['phase'],x['opened_settings']) for x in probe['trials']]==[(kind+'-'+str(i),kind=='unicode') for kind in ['sendkeys','unicode','held-shift'] for i in range(1,4)]
assert probe['instrumented_sha256'].lower()=='75a052d5be803cba84a798be8b4a6367470d57de34ac383be1b0bcbb599470bc'
raw=[json.loads(x) for x in (root/'native/raw-key-events.jsonl').read_text().splitlines()]
decoded=[json.loads(x) for x in (root/'native/decoded-key-events.jsonl').read_text().splitlines()]
by_raw=defaultdict(list);by_dec=defaultdict(list)
for x in raw:assert x['tag']=='[DEBUG-keyS]';by_raw[x['phase']].append(x)
for x in decoded:assert x['tag']=='[DEBUG-keyS]';by_dec[x['phase']].append(x)
matrix=[]
for kind,char,opens in [('sendkeys',115,False),('unicode',83,True),('held-shift',115,False),('caps-on-physical',83,True)]:
 for i in range(1,4):
  phase=f'{kind}-{i}';r=by_raw[phase];d=[x for x in by_dec[phase] if x['event']=='press']
  assert len(r)==len(d)==1 and r[0]['key']['Char']==char and d[0]['string']==chr(char),phase
  assert d[0]['key']['Text']==chr(char) and not d[0]['terminal_focused'] and d[0]['transition']==0
  matrix.append({'phase':phase,'recorded_char':chr(char),'decoded_text':d[0]['string'],'expected_setting_match':opens})
assert not by_raw['caps-off-shift-1'] and not by_dec['caps-off-shift-1']
assert result['result']==audit['result']=='passed' and result['sequence']==269 and len(result['trials'])==6
assert all(x['opened_settings'] and x['capsLock'] and not x['used_shift'] for x in result['trials'])
assert result['caps_before']['capsLock'] and not result['caps_off']['capsLock'] and result['caps_restored']['capsLock'] and result['caps_after_control']['capsLock']
assert {result[k]['layout'] for k in ['caps_before','caps_off','caps_restored','caps_after_control']}=={134481924}
assert result['caps_before']['loaded_layouts']==[134481924]
before=result['before'] if isinstance(result['before'],list) else [result['before']];after=result['after'] if isinstance(result['after'],list) else [result['after']]
assert len(before)==len(after)==1;before,after=before[0],after[0]
assert (before['session_id'],before['run_id'],before['backend_id'],before['command'])==(after['session_id'],after['run_id'],after['backend_id'],after['command'])
assert {identity(x) for x in result['before_processes']}=={identity(x) for x in result['after_processes']}
observations=read('native/native-observations.json');assert len(observations)==5
pane=None
for o in observations:
 assert o['session_id']==before['session_id'] and o['run_id']==before['run_id']
 if pane is None:pane=o['pane']
 assert pane==o['pane']
 assert [int(x) for x in re.findall(r'^history line (\d+)\s*$',o['capture_text'],re.M)]==list(range(1,41))
 lines=[line.strip() for line in o['capture_text'].splitlines() if line.strip()]
 want='> S' if o['label'].startswith('original-') else '>'
 assert lines[-1]==want,o['label']
responses=sorted((root/'native/desktop/responses').glob('*.json'));done=sorted((root/'native/desktop/done').glob('*.json'))
assert [p.name for p in responses]==[f'{i:03d}.json' for i in range(1,270)]==[p.name for p in done]
kinds=Counter();captures={}
for rp,cp in zip(responses,done):
 r=read(str(rp.relative_to(root)));c=read(str(cp.relative_to(root)));assert r['ok'] and r['command']==c,rp.name
 kinds[c['kind']]+=1
 if c['kind']=='capture':captures[c['label']]=r['state']
assert read('native/desktop/continuation-ready.json')['pending']==['228.json']
assert read('native/desktop/responses/228.json')['command']=={'kind':'keys','keys':'{ESC}'}
assert read('native/desktop/driver-continuation-error.json')['error']=='Foreground changed; input refused'
for name in ['clipboard-first-driver-restored.json','clipboard-prior-continuation-restored.json','clipboard-restored.json']:assert read('native/desktop/'+name)['restored']
assert audit['private_images_alive']==0 and audit['remaining_sessions']=='[]' and len(audit['original_processes'])==5 and all(x['original_ended'] for x in audit['original_processes'])
ready=read('native/desktop/ready.json');expected={identity(x) for x in result['before_processes']+result['ui_processes']+[ready]}
assert len(expected)==5 and expected=={identity(x) for x in audit['original_processes']}
assert len(audit['tasks_absent'])==10
assert audit['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert audit['caps_before'] and audit['caps_verified_before_window_close']
assert audit['clipboard']['restored'] and audit['clipboard']['nontext_content_verified'] is False
pngs=sorted((root/'native/desktop').glob('*.png'));assert len(pngs)==13
for p in pngs:
 width,height=struct.unpack('>II',p.read_bytes()[16:24]);assert [width,height]==captures[p.stem]['rect'][2:]
summary={'result':'bounded_uppercase_text_routing_verified','artifacts':len(manifest['files']),'unchanged_source_fingerprints':len(source['files']),'receipts':len(responses),'action_kinds':dict(kinds),'baseline_trials':base['trials'],'boundary_matrix':matrix,'raw_records':len(raw),'decoded_records':len(decoded),'caps_off_shift_records_observed':len(by_raw['caps-off-shift-1']),'green_caps_on_trials':len(result['trials']),'complete_numbered_captures':len(observations),'original_processes_ended':len(expected),'tasks_absent':len(audit['tasks_absent']),'caps_restored':True,'private_images_alive':0,'unmodified_pngs':len(pngs),'visual_gap':'caps-on-physical-1 PNG remains manager while adjacent UIA reports settings; not used as settings-image acceptance','open_gate':'Caps-off Shift path absent from the recorded console/decode logs; full IME/physical keyboard acceptance remains open'}
(root/'checks-summary.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ['baseline_trials','boundary_matrix']},ensure_ascii=False))
