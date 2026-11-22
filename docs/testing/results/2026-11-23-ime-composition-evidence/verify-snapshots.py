from pathlib import Path
import json,re
root=Path('/private/tmp/asd-ime-20261123')
def read(name):return json.loads((root/name).read_text(encoding='utf-8-sig'))
def response(n):return read(f'desktop/responses/{n:03}.json')
def field(n):
 text=response(n)['state']['text'];m=re.search(r'名称（可空）[^\r\n]*\r?\n([^\r\n]*)',text)
 assert m is not None,n
 return m[1]
assert '你好' not in field(32)
assert response(32)['state']['text'].count('你好')==1  # original whole-screen predicate included the dimmed background
assert field(34).count('你好')==1
assert field(36)==field(34)==field(38)
assert '新建终端' in response(38)['state']['text']
assert '新建终端' not in response(40)['state']['text']
assert response(1)['state']['capsLock'] is True
assert response(2)['capsLock'] is False
assert response(42)['capsRestored'] is True and response(42)['state']['capsLock'] is True
preedit=read('preview-preedit.json')['snapshot']['text'];confirmed=read('preview-confirmed.json')['snapshot']['text']
assert preedit.rstrip('\n')=="shell> Set-Content ime-confirm.txt '"
assert confirmed.rstrip('\n')=="shell> Set-Content ime-confirm.txt '你好"
assert confirmed.count('你好')==1
assert read('preview-cancel-preedit.json')['snapshot']['text']==confirmed
assert read('preview-cancelled.json')['snapshot']['text']==confirmed
rows=[read(f'desktop/responses/{n:03}.json') for n in range(1,43)]
assert all(r['ok'] for r in rows)
assert read('terminal-ime-result.json')['text']==read('control-form-ime-result.json')['direct_control']=='你好'
for name in ('ime-confirm.txt','ime-control.txt'):
 assert (root/name).read_text(encoding='utf-8-sig').strip()=='你好',name
print('42 actual desktop responses, scoped form preedit/confirm/cancel, native terminal snapshots and Caps restoration verified')
