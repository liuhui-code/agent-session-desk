"""Compare visible SGR roles emitted by pinned source and product formatters."""
import json
from pathlib import Path
import re
ROOT=Path('/private/tmp/asd-footer-input-20261123')
SGR=re.compile(r'\x1b\[([0-9;]*)m')
def cells(text):
    fg=bg=None;bold=False;offset=0;result=[]
    def append(segment):
        for char in segment:
            result.append({'text':char,'background':bg,'foreground':fg if not char.isspace() else None,'bold':bold if not char.isspace() else False})
    for match in SGR.finditer(text):
        append(text[offset:match.start()]);codes=[int(v or 0) for v in match.group(1).split(';')];i=0
        while i<len(codes):
            code=codes[i]
            if code==0:fg=bg=None;bold=False
            elif code==1:bold=True
            elif code==22:bold=False
            elif code==39:fg=None
            elif code==49:bg=None
            elif code in (38,48) and i+1<len(codes):
                color=None
                if codes[i+1]==2 and i+4<len(codes):color=codes[i+2:i+5];i+=4
                elif codes[i+1]==5 and i+2<len(codes):color=['idx',codes[i+2]];i+=2
                if code==38:fg=color
                else:bg=color
            i+=1
        offset=match.end()
    append(text[offset:]);return result
source=json.loads((ROOT/'upstream-reference.json').read_text())
product=json.loads((ROOT/'product-reference.json').read_text())
assert source['default_footer']=='full' and len(source['frames'])==len(product['frames'])==24
assert len(source['paste_cases'])==10 and all(not c['visible'] and c['offset']==0 and c['all_after']==c['all'] for c in source['paste_cases'])
roles={}
for role in ['role_full','role_compact']:
    a,b=cells(source[role]),cells(product[role]);assert a==b,(role,a,b);roles[role]=len(a)
source_terminal=next(f for f in source['frames'] if f['columns']==120 and f['mode']=='terminal')['content'].split('\n')[1]
badge=cells(product['role_badge']);assert cells(source_terminal)[:len(badge)]==badge;roles['role_badge']=len(badge)
result={'result':'passed','source_default_footer':source['default_footer'],'source_frame_count':len(source['frames']),'product_frame_count':len(product['frames']),'paste_cases':len(source['paste_cases']),'matching_role_cells':roles,'scope':'Visible text/background/foreground/bold in formatter roles only; localized footer content and Tab controls differ, not full-frame equality'}
(ROOT/'role-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
