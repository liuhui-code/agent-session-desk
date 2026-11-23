from pathlib import Path
import json,subprocess
out=Path('/private/tmp/asd-custom-groups-20261123')
results=[]
for name,cmd in [('go-test-all',['go','test','./...']),('go-test-race',['go','test','-race','./...']),('go-vet',['go','vet','./...']),('go-mod-verify',['go','mod','verify'])]:
    old=out/(name+'.log')
    if old.exists():(out/(name+'-before-observer-fix.log')).write_bytes(old.read_bytes())
    r=subprocess.run(cmd,capture_output=True)
    old.write_bytes(r.stdout+r.stderr)
    results.append({'command':cmd,'exit_code':r.returncode})
    (out/'check-exits.json').write_text(json.dumps(results,indent=2)+'\n')
    if r.returncode:raise SystemExit(r.returncode)
print(json.dumps(results))
