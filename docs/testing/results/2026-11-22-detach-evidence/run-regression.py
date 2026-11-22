import subprocess,json
from pathlib import Path
root=Path('/private/tmp/asd-detach-20261122')
def request(i,c):
 r=subprocess.run(['python3',str(root/'request.py'),f'{i:04d}',json.dumps(c,ensure_ascii=False)],capture_output=True,text=True)
 if r.returncode: raise RuntimeError(r.stdout+r.stderr)
 d=json.loads((root/f'response-{i:04d}.json').read_text(encoding='utf-8-sig'))
 if not d.get('ok'): raise RuntimeError(str(d))
 return d
rows=[]
for start in (40,45):
 request(start,{'kind':'keys','keys':'{UP}{ENTER}'})
 before=request(start+1,{'kind':'capture','label':f'direct-english-before-{start}'})
 assert 'ASD native Unicode fixture ready' in before['state']['text'] and 'Agent Session Desk' not in before['state']['text']
 request(start+2,{'kind':'keys','keys':'^b'})
 request(start+3,{'kind':'physical','text':'d'})
 after=request(start+4,{'kind':'capture','label':f'direct-english-after-{start}'})
 assert 'PS C:' in after['state']['text']
 rows.append({'path':'direct psmux','before_command':start+1,'after_command':start+4,'physical_d':True,'passed':True})
(root/'direct-regression.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows))
