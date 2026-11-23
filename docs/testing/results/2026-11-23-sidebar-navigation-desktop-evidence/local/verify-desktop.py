from pathlib import Path
from collections import Counter
import hashlib,json,re,struct

root=Path('docs/testing/results/2026-11-23-sidebar-navigation-desktop-evidence')
def normalize(value):
 if isinstance(value,dict):
  if set(value)=={'value','Count'}:
   assert isinstance(value['value'],list) and len(value['value'])==value['Count']
   return normalize(value['value'])
  return {k:normalize(v) for k,v in value.items()}
 if isinstance(value,list):return [normalize(x) for x in value]
 return value
def read(name):return normalize(json.loads((root/name).read_text()))
def identity(x):
 assert isinstance(x['pid'],int) and isinstance(x['filetime'],int) and isinstance(x['path'],str)
 return x['pid'],x['filetime'],x['path'].lower()
manifest=read('manifest.json')
for entry in manifest['files']:
 data=(root/entry['path']).read_bytes()
 assert len(data)==entry['archived_bytes'] and hashlib.sha256(data).hexdigest()==entry['archived_sha256'],entry['path']
 if not entry['path'].endswith('.png'):
  assert b'<LOCAL_TEST_HOME>' not in data and b'<WINDOWS_TEST_USER>' not in data,entry['path']
source=read('local/source-before.json');after=read('local/source-after.json')
assert source==after
for item in source['files']:assert hashlib.sha256(Path(item['path']).read_bytes()).hexdigest()==item['sha256'],item['path']
receipt=read('native/receipt.json');flow=read('native/flow-result.json');audit=read('native/final-audit.json')
assert receipt['app_sha256'].lower()==source['binary_sha256']=='f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277'
assert read('native/setup-outcome.json')['is_administrator'] is False
assert read('native/setup-outcome.json')['session']==2
assert flow['result']==audit['result']=='passed'
assert flow['commands']==283 and flow['preference_snapshots']==17
stable=lambda x:(x['session_id'],x['run_id'],x['backend_id'],x['title'],x['command'])
assert len(flow['before'])==len(flow['after'])==2
assert {stable(x) for x in flow['before']}=={stable(x) for x in flow['after']}
assert {identity(x) for x in flow['before_processes']}=={identity(x) for x in flow['after_processes']}
assert len({x['pid'] for x in flow['before_processes']})==2
known={x['title']:x for x in flow['before']};pids={x['label']:x['pid'] for x in flow['before_processes']}
observations=read('native/native-observations.json');assert len(observations)==13
panes={};pending=0
for o in observations:
 s=next(x for x in known.values() if x['session_id']==o['session_id']);assert o['run_id']==s['run_id']
 assert panes.setdefault(o['session_id'],o['pane'])==o['pane']
 ids=[int(x) for x in re.findall(r'^history line (\d+)\s*$',o['capture_text'],re.M)]
 assert ids==list(range(1,41)),o['label']
 label='NAV_A' if s['title']=='Copy A' else 'NAV_B'
 opposite='NAV_B' if label=='NAV_A' else 'NAV_A'
 assert not re.search(r'^'+opposite+r':\d+\s*$',o['capture_text'],re.M)
 if o['label'].endswith('explicit-enter'):
  outputs=re.findall(r'^'+label+r':(\d+)\s*$',o['capture_text'],re.M)
  assert outputs==[str(pids[s['title']])],o['label']
 else:
  lines=[line.strip() for line in o['capture_text'].splitlines() if line.strip()]
  assert lines[-1]=="> Write-Output ('"+label+":'+$PID)",o['label']
  assert not re.search(r'^'+label+r':\d+\s*$',o['capture_text'],re.M)
  pending+=1
assert pending==11
responses=sorted((root/'native/desktop/responses').glob('*.json'));done=sorted((root/'native/desktop/done').glob('*.json'))
assert [p.name for p in responses]==[f'{i:03d}.json' for i in range(1,284)]
assert [p.name for p in done]==[p.name for p in responses]
kinds=Counter();captures={}
for response,command in zip(responses,done):
 r=read(str(response.relative_to(root)));c=read(str(command.relative_to(root)))
 assert r['ok'] is True and r['command']==c,response.name
 kinds[c['kind']]+=1
 if c['kind']=='capture':captures[c['label']]=r['state']
preferences=read('native/preferences.json');assert len(preferences)==17
assert {x['density'] for x in preferences}=={'full','compact','minimal','auto'}
for p in preferences:
 assert hashlib.sha256(p['text'].encode()).hexdigest()==p['sha256'].lower(),p['label']
 assert p['text'].splitlines()==['schema_version = 2',f'sidebar_density = "{p["density"]}"',f'sidebar_mode = "{p["mode"]}"',f'default_group_collapsed = {str(p["collapsed"]).lower()}']
assert preferences[-1]['text']==(root/'native/ui-final.toml').read_text()
assert flow['legacy_read_hash'].lower()==hashlib.sha256((root/'native/ui-v1-before.toml').read_bytes()).hexdigest()
assert next(x for x in preferences if x['label']=='UI-reopen-retains-fold-density-mode')['collapsed'] is True
pngs=sorted((root/'native/desktop').glob('*.png'));assert len(pngs)==11
for png in pngs:
 assert png.stem in captures,png.name
 width,height=struct.unpack('>II',png.read_bytes()[16:24]);assert [width,height]==captures[png.stem]['rect'][2:]
 assert width in [1113,600,473] and height in [657,539]
for name,width in [('grouped-auto-63-public',63),('grouped-auto-49-public',49)]:
 rows=[r for r in captures[name]['text'].splitlines() if 'Copy A ×' in r and 'Copy B ×' in r]
 assert len(rows)==1 and len(rows[0])==width,name
assert audit['private_images_alive']==0 and audit['remaining_sessions']=='[]'
assert len(audit['original_processes'])==5 and all(x['original_ended'] for x in audit['original_processes'])
ready=read('native/desktop/ready.json')
assert 'Consolas' in ready['fontNames']
assert json.loads((root/'local/catalog-check.log').read_text().splitlines()[0])['cases']==339
assert 'all files <= 500 lines' in (root/'local/source-limits.log').read_text()
expected={identity(x) for x in flow['before_processes']+flow['ui_processes']+[ready]}
assert len(expected)==5 and {identity(x) for x in audit['original_processes']}==expected
assert len(audit['stops'])==2
assert set(audit['tasks_absent'])=={'AsdSidebarNavigationSetup20261123','AsdSidebarNavigationDesktop20261123'}
assert audit['settings_sha256'].lower()==receipt['settings_before'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert audit['clipboard']['restored'] and audit['clipboard']['text_restored'] and audit['clipboard']['format_names_restored']
assert audit['clipboard']['nontext_content_verified'] is False
failures={name:read('native/'+name)['error'] for name in ['flow-failure-columns.json','flow-failure-sendkeys-uppercase.json','flow-failure-active-marker.json','flow-failure-reopen-active-marker.json']}
assert failures['flow-failure-sendkeys-uppercase.json']=='Expected screen not observed: 侧栏密度'
summary={'result':'passed_for_bounded_scope','source_fingerprints':len(source['files']),'artifacts':len(manifest['files']),'action_receipts':len(responses),'action_kinds':dict(kinds),'native_captures':len(observations),'complete_numbered_captures':len(observations),'pending_drafts_checked':pending,'preference_snapshots':len(preferences),'unmodified_pngs':len(pngs),'original_processes_ended':len(expected),'private_images_alive':0,'preserved_failure_errors':failures,'unresolved':'SendKeys uppercase S did not open settings; Unicode S worked. Physical Shift+S/IME/framework boundary unproven; not a full keyboard acceptance.'}
(root/'checks-summary.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(summary,ensure_ascii=False))
