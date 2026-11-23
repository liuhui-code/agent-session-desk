"""Compare actual source/product card surfaces and divider cell roles."""
from pathlib import Path
import json,re,unicodedata,sys
R=Path(sys.argv[1]) if len(sys.argv)>1 else Path('/private/tmp/asd-main-parity-20261123')
SGR=re.compile(r'\x1b\[([0-9;]*)m')
def width(c):
    if unicodedata.combining(c):return 0
    return 2 if unicodedata.east_asian_width(c) in ('W','F') else 1
def cells(line):
    fg=bg=None;offset=0;out=[]
    def append(s):
        for ch in s:
            assert ch not in '\x1b\n\r\t',repr(ch)
            out.extend([{'text':ch,'foreground':fg,'background':bg}]*width(ch))
    for m in SGR.finditer(line):
        append(line[offset:m.start()]);cc=[int(v or 0) for v in m.group(1).split(';')];i=0
        while i<len(cc):
            c=cc[i]
            if c==0:fg=bg=None
            elif c==39:fg=None
            elif c==49:bg=None
            elif c in (38,48) and i+1<len(cc):
                v=None
                if cc[i+1]==2 and i+4<len(cc):v=cc[i+2:i+5];i+=4
                elif cc[i+1]==5 and i+2<len(cc):v=['idx',cc[i+2]];i+=2
                if c==38:fg=v
                else:bg=v
            i+=1
        offset=m.end()
    append(line[offset:]);return out
source=json.loads((R/'upstream-main-reference.json').read_text())
current=json.loads((R/'product-main-reference.json').read_text())
before=json.loads((R/'product-main-before.json').read_text())
assert len(source['frames'])==len(current['cards'])==len(before['cards'])==27
checks=0;holes=0;highlight=0
for s,a,b in zip(source['frames'],before['cards'],current['cards']):
    for key in ('columns','level','mode'):assert s[key]==a[key]==b[key]
    ss,aa,bb=[v['content'].splitlines() for v in (s,a,b)];assert len(ss)==len(aa)==len(bb)==3
    for sr,ar,br in zip(ss,aa,bb):
        sc,ac,bc=map(cells,(sr,ar,br));assert len(sc)==len(ac)==len(bc)==s['columns']
        for x,y,z in zip(sc,ac,bc):
            assert x['background']==z['background'],(s,z)
            if x['background']!=y['background']:holes+=1
            if s['mode']!='plain':assert x['background']==[36,40,59];highlight+=1
            checks+=1
assert checks==3591 and holes==108 and highlight==2394
accent_deltas=[];dividers=[];old_divider_differences=[]
for s in source['dividers']:
    mode=s['mode'];p=next(f for f in current['frames'] if f['columns']==120 and f['rows']==32 and f['mode']==mode)
    sc=cells(s['content'].splitlines()[0])[s['left']+1]
    g=p['geometry'];pc=cells(p['content'].splitlines()[2])[g['list_width']+1]
    assert sc['text']==pc['text']==('┃' if mode in ('hover','drag') else '│')
    expected=dict(sc)
    if sc['foreground']==[121,162,247] and pc['foreground']==[122,162,247]:expected['foreground']=pc['foreground'];accent_deltas.append(mode)
    assert expected==pc,(mode,sc,pc)
    old=next(f for f in before['frames'] if f['columns']==120 and f['rows']==32 and f['mode']==mode)
    oc=cells(old['content'].splitlines()[2])[old['geometry']['list_width']+1]
    if oc!=pc:old_divider_differences.append(mode)
    dividers.append({'mode':mode,'source_cell':sc,'product_cell':pc,'source_list_width':s['left'],'product_list_width':g['list_width']})
assert accent_deltas==['terminal','hover','drag']
assert old_divider_differences==['terminal','hover','drag']
for f in current['frames']:
    rr=f['content'].split('\n');assert len(rr)==f['rows'] and f['mouse_mode']==2
    assert all(len(cells(v))==f['columns'] for v in rr)
    assert all(v['text']=='─' for v in cells(rr[-3]))
assert len(current['frames'])==72
result={'result':'passed','source_card_pairs':27,'card_background_cells':checks,'highlight_cells':highlight,'before_holes':holes,'after_holes':0,'divider_roles':dividers,'before_divider_differences':old_divider_differences,'known_accent_serialization_delta_modes':accent_deltas,'product_main_frames':72,'static_sizes':sorted({(v['columns'],v['rows']) for v in current['frames']}),'scope':'Actual renderers with synthetic state; background and specific divider cells only, no whole-frame equality, Windows or real-agent claim. Source terminal divider case uses explicit noncompact 35% while product input mode uses 18%; prior actual source compact frames remain separate.'}
(R/'role-comparison.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='divider_roles'},ensure_ascii=False))
