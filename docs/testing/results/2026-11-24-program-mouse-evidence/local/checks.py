from pathlib import Path
import subprocess,json,datetime
out=Path('/private/tmp/asd-program-mouse-20261124');results=[]
for name,args in [('go-test',['go','test','./...']),('go-race',['go','test','-race','./...']),('go-vet',['go','vet','./...']),('go-mod-verify',['go','mod','verify']),('source-limits',['python3','-B','tools/ci/check_source_limits.py'])]:
 start=datetime.datetime.now(datetime.timezone.utc).isoformat()
 with (out/(name+'.log')).open('w') as f:r=subprocess.run(args,stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(check=name,command=args,exit_code=r.returncode,start_utc=start,end_utc=datetime.datetime.now(datetime.timezone.utc).isoformat()))
 (out/'checks.json').write_text(json.dumps(results,indent=2))
 print(name,r.returncode,flush=True)
 if r.returncode:raise SystemExit(r.returncode)
