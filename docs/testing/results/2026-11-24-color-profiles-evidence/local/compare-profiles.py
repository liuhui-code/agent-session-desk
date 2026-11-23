from pathlib import Path
import json,re
ROOT=Path('/private/tmp/asd-color-profiles-20261124')
ESC=re.compile(r'\x1b(?:\[([0-?]*)([ -/]*)([@-~])|\][^\x07]*(?:\x07|\x1b\\)|[@-_])')
def cells(text):
 fg=bg=None;bold=italic=underline=False;offset=0;result=[]
 def append(segment):
  for char in segment:
   if char not in '\r\n':result.append(dict(text=char,foreground=fg,background=bg,bold=bold,italic=italic,underline=underline))
 for m in ESC.finditer(text):
  append(text[offset:m.start()])
  if m.group(3)=='m':
   codes=[int(v or 0) for v in m.group(1).split(';')];i=0
   while i<len(codes):
    c=codes[i]
    if c==0:fg=bg=None;bold=italic=underline=False
    elif c==1:bold=True
    elif c==22:bold=False
    elif c==3:italic=True
    elif c==23:italic=False
    elif c==4:underline=True
    elif c==24:underline=False
    elif c==39:fg=None
    elif c==49:bg=None
    elif 30<=c<=37:fg=['palette',c-30]
    elif 90<=c<=97:fg=['palette',c-90+8]
    elif 40<=c<=47:bg=['palette',c-40]
    elif 100<=c<=107:bg=['palette',c-100+8]
    elif c in (38,48) and i+1<len(codes):
     color=None
     if codes[i+1]==2 and i+4<len(codes):color=['rgb',*codes[i+2:i+5]];i+=4
     elif codes[i+1]==5 and i+2<len(codes):color=['palette',codes[i+2]];i+=2
     if c==38:fg=color
     else:bg=color
    i+=1
  offset=m.end()
 append(text[offset:]);return result

def probes(case):
 data=cells(case['actual_program_output']);plain=''.join(v['text'] for v in data);result={}
 for v in case['labels']:
  marker=v['tag']+'|';i=plain.index(marker)+len(marker);j=plain.index('|',i)
  value=data[i:j]
  assert value and ''.join(c['text'] for c in value)==''.join(c['text'] for c in cells(v['rendered'])),v
  result[(v['theme'],v['kind'],v['role'])]=value
 return result

source=json.loads((ROOT/'source-reference.json').read_text());product=json.loads((ROOT/'product-reference.json').read_text())
assert len(source)==len(product)==4
results=[]
for s,p in zip(source,product):
 assert s['profile']==p['profile']
 left,right=probes(s),probes(p);assert set(left)==set(right) and len(left)==30
 cases=[]
 for key,a in sorted(left.items()):
  b=right[key];cases.append(dict(theme=key[0],kind=key[1],role=key[2],source=a,product=b,matches=a==b))
 results.append(dict(profile=s['profile'],cases=cases,matches=sum(v['matches'] for v in cases),total=len(cases)))
(ROOT/'profile-comparison.json').write_text(json.dumps(dict(result='observed_actual_renderer_profiles',profiles=results,scope='normalized SGR cells from memory output of real v1/v2 renderers; palette indices are compared, not terminal-specific RGB; not a real terminal, Windows or full UI acceptance'),indent=2)+'\n')
print(json.dumps([dict(profile=r['profile'],matches=r['matches'],total=r['total'],different=[v['theme']+'/'+v['role'] for v in r['cases'] if not v['matches']]) for r in results]))
