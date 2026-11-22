import json,subprocess,sys
from pathlib import Path
root=Path('/private/tmp/asd-detach-20261122')
id=sys.argv[1]; command=json.loads(sys.argv[2]); payload=json.dumps(command,ensure_ascii=False).replace("'","''")
source=f'''$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/detach-diagnosis'
$p=Join-Path $b 'commands/{id}.json';$r=Join-Path $b 'responses/{id}.json'
if((Test-Path $p)-or(Test-Path $r)){{throw 'Command ID already exists'}}
[IO.File]::WriteAllText(($p+'.tmp'),'{payload}',[Text.UTF8Encoding]::new($false));Move-Item ($p+'.tmp') $p
$until=[DateTime]::UtcNow.AddSeconds(8)
while(!(Test-Path $r)-and[DateTime]::UtcNow-lt$until){{Start-Sleep -Milliseconds 100}}
if(Test-Path $r){{Get-Content $r -Raw -Encoding UTF8}}else{{if(Test-Path (Join-Path $b 'driver-error.json')){{Get-Content (Join-Path $b 'driver-error.json') -Raw -Encoding UTF8}}else{{throw 'Desktop response pending'}}}}
'''
p=root/f'req-{id}.ps1';p.write_text(source)
r=subprocess.run(['python3','/private/tmp/asd-remote.py',str(p),str(root/f'response-{id}.json')],capture_output=True,text=True)
if r.returncode:print(r.stdout,r.stderr);raise SystemExit(r.returncode)
try:
 d=json.loads(r.stdout.lstrip('\ufeff')); state=d.get('state',{});print(json.dumps({k:v for k,v in d.items() if k!='state'},ensure_ascii=False));print(json.dumps(state,ensure_ascii=False)[:6200])
except Exception:print(r.stdout)
