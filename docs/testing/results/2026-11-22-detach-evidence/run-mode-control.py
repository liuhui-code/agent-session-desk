import json,subprocess
from pathlib import Path
b=Path('/private/tmp/asd-detach-20261122')
def req(i,c):
 r=subprocess.run(['python3',str(b/'request.py'),f'{i:04d}',json.dumps(c,ensure_ascii=False)],capture_output=True,text=True)
 if r.returncode:raise RuntimeError(r.stdout+r.stderr)
 d=json.loads((b/f'response-{i:04d}.json').read_text(encoding='utf-8-sig'))
 assert d.get('ok'),d
 return d
req(70,{'kind':'keys','keys':'{ENTER}'})
req(71,{'kind':'physical','text':chr(16)})
req(72,{'kind':'keys','keys':'^b'});req(73,{'kind':'physical','text':'d'})
red=req(74,{'kind':'capture','label':'managed-chinese-red'})
assert 'Agent Session Desk' not in red['state']['text']
req(75,{'kind':'keys','keys':'{ESC}'})
req(76,{'kind':'physical','text':chr(16)})
req(77,{'kind':'physical','text':'d'})
green=req(78,{'kind':'capture','label':'managed-english-green'})
assert 'Agent Session Desk' in green['state']['text'] and '已返回管理器' in green['state']['text']
(b/'mode-control.json').write_text(json.dumps({'same_attach_client':True,'chinese_physical_d_did_not_detach':True,'then_cancel_candidate_switch_english_physical_d_only_detached':True,'prefix_not_resent':True,'red_command':74,'green_command':78},indent=2)+'\n')
print('Same-client input-mode control passed: no prefix resend, no Unicode d fallback.')
