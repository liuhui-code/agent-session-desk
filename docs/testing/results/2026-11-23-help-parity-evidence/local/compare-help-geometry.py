"""Compare actual reference and product border geometry without executing either."""
from pathlib import Path
import json
import re
root=Path('/private/tmp/asd-help-parity-20261123')
upstream=json.loads((root/'upstream-help-frames.json').read_text())['frames']
product=json.loads((root/'product-help-frames.json').read_text())
ansi=re.compile(r'\x1b\[[0-?]*[ -/]*[@-~]')
def geometry(frame):
    lines=[ansi.sub('',line) for line in frame['content'].split('\n')]
    top=next((i,line) for i,line in enumerate(lines) if '╭' in line)
    bottom=next((i,line) for i,line in enumerate(lines) if '╰' in line)
    return {'x':top[1].index('╭'),'y':top[0],
            'width':top[1].index('╮')-top[1].index('╭')+1,'height':bottom[0]-top[0]+1}
result=[]
for source,current in zip(upstream,product):
    assert (source['columns'],source['rows'],source['mode'])==(current['columns'],current['rows'],current['mode'])
    result.append({'columns':source['columns'],'rows':source['rows'],'mode':source['mode'],
                   'upstream':geometry(source),'product':geometry(current)})
(root/'help-geometry.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
