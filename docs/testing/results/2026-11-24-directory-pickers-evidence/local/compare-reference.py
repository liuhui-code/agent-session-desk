from pathlib import Path
import json,re
ROOT=Path('/private/tmp/asd-directory-picker-20261124')
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


source=json.loads((ROOT/'source-reference.json').read_text());product=json.loads((ROOT/'product-frames.json').read_text())
profiles={'TrueColor':'truecolor','ANSI256':'ansi256','ANSI':'ansi','Ascii':'ascii','ASCII':'ascii'}
results=[]
for f in product:
 if f['width']!=120 or f['height']!=40:continue
 p=profiles[f['profile']]
 ref=next(x for x in source['frames'] if x['profile']==p and x['height']==40 and x['theme']==f['theme'] and x['selected']==3)
 a=cells(ref[f['mode']]);b=cells(f['content']);at=''.join(v['text']for v in a);bt=''.join(v['text']for v in b)
 candidates=['/projects/alpine','/work/中文 空格','/work/other','/work/five']
 candidates.append('/projects/Alpha' if f['mode']=='launch' else '/work/six')
 for path in candidates:
  assert path in at and path in bt,(f['mode'],p,path)
  left=a[at.index(path):at.index(path)+len(path)];right=b[bt.index(path):bt.index(path)+len(path)]
  results.append(dict(mode=f['mode'],theme=f['theme'],profile=p,path=path,matches=left==right,source=left,product=right))
report=dict(scope='fixed component style cells only; localized footer, screen placement, widths and whole-frame acceptance are separate',cases=len(results),matching=sum(x['matches']for x in results),results=results)
(ROOT/'style-comparison.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(dict(cases=report['cases'],matching=report['matching'],different=[(x['mode'],x['theme'],x['profile'],x['path'])for x in results if not x['matches']])))
assert report['cases']==report['matching']==80
