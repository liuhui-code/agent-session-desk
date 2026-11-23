from pathlib import Path
import json,subprocess
p=Path('/private/tmp/asd-groups-desktop-20261123');results=[]
for name,args in [('go-test',['go','test','./...']),('go-race',['go','test','-race','./...']),('go-vet',['go','vet','./...']),('go-mod-verify',['go','mod','verify'])]:
 r=subprocess.run(args,capture_output=True);(p/(name+'.log')).write_bytes(r.stdout+r.stderr);results.append({'command':args,'exit_code':r.returncode});(p/'check-exits.json').write_text(json.dumps(results,indent=2)+'\n')
 if r.returncode:raise SystemExit(r.returncode)
print(json.dumps(results))
