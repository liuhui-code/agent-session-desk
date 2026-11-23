"""Compare visible SGR roles emitted by pinned source and product formatters."""
import json
from pathlib import Path
import re
ROOT=Path('/private/tmp/asd-source-colors-20261124')
SGR=re.compile(r'\x1b\[([0-9;]*)m')
def cells(text):
    fg=bg=None;bold=False;offset=0;result=[]
    def append(segment):
        for char in segment:
            result.append({'text':char,'background':bg,'foreground':fg,'bold':bold})
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
source=json.loads((ROOT/'source-reference.json').read_text())
product=json.loads((ROOT/'product-reference.json').read_text())
assert len(source['frames'])==len(product['frames'])==144
assert source['normalization']==product['normalization']
results=[]
for left,right in zip(source['roles'],product['roles']):
 assert left['theme']==right['theme'] and left['author_dim']==right['author_dim']
 for role in ['full','compact','curated']:
  a,b=cells(left[role]),cells(right[role]);assert len(a)==len(b)
  deltas=[dict(cell=i,source=l,product=r) for i,(l,r) in enumerate(zip(a,b)) if l!=r]
  results.append(dict(theme=left['theme'],role=role,matches=not deltas,cells=len(a),differences=deltas))

left_palette={(v['theme'],v['role']):v for v in source['palette']}
right_palette={(v['theme'],v['role']):v for v in product['palette']}
assert set(left_palette)==set(right_palette) and len(left_palette)==20
palette=[]
for key,left in sorted(left_palette.items()):
 right=right_palette[key]
 assert left['authored']==right['authored']
 a,b=cells(left['rendered']),cells(right['rendered'])
 palette.append(dict(theme=key[0],role=key[1],authored=left['authored'],source=a,product=b,matches=a==b))
assert all(v['matches'] for v in palette) and all(v['matches'] for v in results)
result=dict(result='passed_exact_truecolor_roles',source_frames=144,product_frames=144,normalization_cases=len(source['normalization']),roles=results,palette=palette,matching_roles=sum(r['matches'] for r in results),total_roles=len(results),matching_palette_roles=sum(r['matches'] for r in palette),scope='20 authored/display roles and 6 combined ASCII style roles only. Localized source/product frames differ; no 16/256-color, full frame, Windows or desktop acceptance. No RGB tolerance.')
(ROOT/'role-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ['roles','palette']}))
