"""Verify archived help evidence; this is not a product execution."""
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import struct
import unicodedata
REPO=Path.cwd()
EVIDENCE=REPO/'docs/testing/results/2026-11-23-help-parity-evidence'
def read(path):
    return json.loads((EVIDENCE/path).read_text())
def sha(data):
    return hashlib.sha256(data).hexdigest()
def rows(value):
    if isinstance(value,dict) and set(value)=={'value','Count'}:
        assert value['Count']==len(value['value']);return value['value']
    assert isinstance(value,list);return value
manifest=read('manifest.json');paths=set()
for entry in manifest['artifacts']:
    path=entry['path'];assert path not in paths and '..' not in Path(path).parts;paths.add(path)
    data=(EVIDENCE/path).read_bytes();assert sha(data)==entry['archived_sha256'] and len(data)==entry['bytes'],path
    assert Path(path).suffix not in ('.exe','.sqlite3','.key','.port')
    if path.endswith('.png'):assert entry['original_sha256']==entry['archived_sha256']
    else:assert not re.search('liu'+' xun yuan|/Users/'+'liuhui',data.decode('utf-8'),re.I),path
fingerprints=read('local/source-fingerprints.json');assert len(fingerprints)==201
for path,expected in fingerprints.items():assert sha((REPO/path).read_bytes())==expected,path
source=read('local/upstream-source.json');assert source['commit']=='80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status']==''
assert len(source['files'])==16
reference=read('local/upstream-help-frames.json');assert reference['tab_selected_all'] and reference['q_closed'] and reference['end_offset']>0
geometry=read('local/help-geometry.json');assert len(geometry)==9
for value in geometry:
    u,p=value['upstream'],value['product']
    assert u['x']==p['x'] and u['width']==p['width']
    if value['columns']==49 and value['mode']=='short':assert p['height']==u['height']-1 and p['y']==u['y']+1
    else:assert u==p

sgr=re.compile(r'\x1b\[([0-9;]*)m')
def background_cells(line):
    background=None;cells=[];offset=0
    def append(text):
        for ch in text:
            width=0 if unicodedata.combining(ch) else (2 if unicodedata.east_asian_width(ch) in ('W','F') else 1)
            cells.extend([background]*width)
    for match in sgr.finditer(line):
        append(line[offset:match.start()]);codes=[int(v or 0) for v in match.group(1).split(';')];i=0
        while i<len(codes):
            code=codes[i]
            if code in (0,49):background=None
            elif code in (38,48) and i+1<len(codes):
                if codes[i+1]==2 and i+4<len(codes):
                    if code==48:background=tuple(codes[i+2:i+5])
                    i+=4
                elif codes[i+1]==5 and i+2<len(codes):
                    if code==48:background=('idx',codes[i+2])
                    i+=2
            i+=1
        offset=match.end()
    append(line[offset:]);return cells

def surface_gaps(frames):
    gaps=0;tested=0
    for frame in frames:
        active=False;x=width=0
        for line in frame['content'].split('\n'):
            plain=sgr.sub('',line)
            if '╭' in plain:
                active=True;x=plain.index('╭');width=plain.index('╮')-x+1;continue
            if '╰' in plain:active=False;continue
            if active:
                colors=background_cells(line)
                for col in range(x+1,x+width-1):
                    tested+=1
                    if col>=len(colors) or colors[col]!=(36,40,59):gaps+=1
    return gaps,tested
frames=read('local/product-help-frames.json');assert len(frames)==9
gaps,cells=surface_gaps(frames);assert gaps==0 and cells>0
before_gaps,_=surface_gaps(read('local/product-help-frames-before-surface.json'));assert before_gaps>0

def native(prefix,hashes_name,binary,top,total):
    result=read(prefix+'/outcome.json')['results'];assert len(result)==1;result=result[0]
    assert (result['top_level_pass'],result['all_pass'])==(top,total)
    assert result['exit_code']==result['fail']==result['skip']==result['remaining_exact_image']==0
    assert result['binary_sha256'].lower()==read('local/'+hashes_name)[binary]
    text=(EVIDENCE/prefix/'ui-tests.log').read_text()
    assert len(re.findall(r'^--- PASS:',text,re.M))==top
    assert len(re.findall(r'^\s*--- PASS:',text,re.M))==total
    assert not re.search(r'^\s*--- (?:FAIL|SKIP):',text,re.M)
native('native-before-surface/native-tests-first','binary-hashes-first.json','help-parity-ui.exe',162,346)
native('native-before-surface/native-tests','binary-hashes-before-surface.json','help-parity-ui.exe',163,347)
native('native/native-tests','binary-hashes.json','help-parity-style-ui.exe',164,348)
assert '--- PASS: TestWorkspaceHelpFillsEveryInteriorCellWithSurface' in (EVIDENCE/'native/native-tests/ui-tests.log').read_text()

def flow(prefix,task_prefix):
    value=read(prefix+'/flow-result.json');assert value['result']=='passed'
    responses=sorted((EVIDENCE/prefix/'desktop/responses').glob('*.json'))
    done=sorted((EVIDENCE/prefix/'desktop/done').glob('*.json'))
    assert [p.stem for p in responses]==[f'{i:03}' for i in range(1,value['commands']+1)]
    assert [p.name for p in responses]==[p.name for p in done]
    kinds=Counter();screenshots=[]
    for response_path,command_path in zip(responses,done):
        response=json.loads(response_path.read_text());command=json.loads(command_path.read_text())
        assert response['ok'] and response['command']==command;kinds[command['kind']]+=1
        if command['kind']=='capture':
            name=command['label']+'.png';data=(EVIDENCE/prefix/'desktop'/name).read_bytes()
            assert data.startswith(b'\x89PNG\r\n\x1a\n')
            assert list(struct.unpack('>II',data[16:24]))==response['state']['rect'][2:];screenshots.append(name)
    for key,count in {'resize':4,'keys':32,'text':17,'capture':10,'paste':4,'finish':1}.items():assert kinds[key]==count
    assert len(screenshots)==10 and kinds['state']>=45
    before={v['session_id']:v for v in rows(value['before'])};after={v['session_id']:v for v in rows(value['after'])}
    assert len(before)==2 and before==after
    processes=rows(value['before_processes']);assert len(processes)==2 and processes==rows(value['after_processes'])
    observations=rows(read(prefix+'/native-observations.json'));assert len(observations)==value['native_captures']==13
    for observation in observations:
        assert observation['complete'] and observation['numbered_ids']==list(range(1,41))
        text=observation['capture_text'];assert [int(v) for v in re.findall(r'^history line (\d+)\s*$',text,re.M)]==list(range(1,41))
        letter=observation['label'][0];owner=processes[0 if letter=='A' else 1]
        sid=next(sid for sid,v in before.items() if v['title']=='Copy '+letter)
        assert observation['session_id']==sid and observation['run_id']==before[sid]['run_id']
        outputs=re.findall(r'^HELP_'+letter+r':(\d+)\s*$',text,re.M)
        if observation['label']==letter+'-explicit-enter':assert outputs==[str(owner['pid'])]
        else:
            assert not outputs
            lines=[line.strip() for line in text.splitlines() if line.strip()]
            expected="> Write-Output ('HELP_"+letter+":'+$PID)"+(' # ?nq' if letter=='A' else '')
            assert lines[-1]==expected
    assert value['preferences_before']==value['preferences_after']
    assert (EVIDENCE/prefix/'ui-v3-before.toml').read_bytes()==(EVIDENCE/prefix/'ui-final.toml').read_bytes()
    audit=read(prefix+'/final-audit.json')
    assert audit['result']=='passed' and audit['private_images_alive']==0 and audit['remaining_sessions']=='[]'
    assert len(audit['stops'])==2 and {(s['session_id'],s['run_id']) for s in audit['stops']}=={(s['session_id'],s['run_id']) for s in before.values()}
    originals=rows(audit['original_processes']);assert len(originals)==5 and all(v['original_ended'] for v in originals)
    assert audit['settings_sha256'].lower()=='1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
    assert sorted(audit['tasks_absent'])==[task_prefix+'Desktop20261123',task_prefix+'Setup20261123']
    clip=audit['clipboard'];assert clip['restored'] and clip['text_restored'] and clip['format_names_restored']
    assert clip['caps_before'] is clip['caps_after'] is True and clip['nontext_content_verified'] is False
    return {'actions':dict(kinds),'commands':value['commands'],'screenshots':len(screenshots),'native_captures':len(observations),'original_processes_ended':len(originals)}
first=flow('native-before-surface','AsdHelpParity')
final=flow('native','AsdHelpParityStyle')
result={'result':'passed','artifacts':len(paths),'source_fingerprints':len(fingerprints),
 'source_product_component_pairs':len(geometry),'final_surface_cells':cells,'surface_gaps_before':before_gaps,'surface_gaps_final':gaps,
 'native_ui_final':[164,348],'desktop_before_surface':first,'desktop_final':final,
 'scope':'Bounded archived evidence consistency; no new product execution or overall acceptance'}
(EVIDENCE/'verification.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False))
