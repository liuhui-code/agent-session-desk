import subprocess,json
from pathlib import Path
root=Path('/private/tmp/asd-detach-20261122')
def req(i,c):
 r=subprocess.run(['python3',str(root/'request.py'),f'{i:04d}',json.dumps(c,ensure_ascii=False)],capture_output=True,text=True)
 if r.returncode:raise RuntimeError(r.stdout+r.stderr)
 d=json.loads((root/f'response-{i:04d}.json').read_text(encoding='utf-8-sig'))
 if not d.get('ok'):raise RuntimeError(str(d))
 return d
rows=[]
for i in (100,105,110):
 req(i,{'kind':'keys','keys':'{ENTER}'})
 before=req(i+1,{'kind':'capture','label':f'candidate-english-attached-{i}'})
 assert 'ASD native Unicode fixture ready' in before['state']['text'] and 'Agent Session Desk' not in before['state']['text']
 req(i+2,{'kind':'keys','keys':'^b'});req(i+3,{'kind':'physical','text':'d'})
 after=req(i+4,{'kind':'capture','label':f'candidate-english-return-{i}'})
 assert 'Agent Session Desk' in after['state']['text'] and '已返回管理器' in after['state']['text']
 rows.append({'path':'ASD manager','before_command':i+1,'after_command':i+4,'physical_d':True,'passed':True})
(root/'candidate-regression.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows))
