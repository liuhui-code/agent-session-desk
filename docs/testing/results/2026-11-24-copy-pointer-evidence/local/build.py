from pathlib import Path
import os,subprocess,json,hashlib
out=Path('/private/tmp/asd-copy-pointer-20261124');env=dict(os.environ,GOOS='windows',GOARCH='amd64',CGO_ENABLED='0');hashes={}
checks=json.loads((out/'checks.json').read_text());assert len(checks)==5 and all(c['exit_code']==0 for c in checks)
for package in ['terminal','ui']:
 name='copy-pointer-'+package+'.exe';subprocess.run(['go','test','-c','-o',str(out/name),'./internal/'+package],env=env,check=True);hashes[name]=hashlib.sha256((out/name).read_bytes()).hexdigest()
name='asd-copy-pointer.exe';subprocess.run(['go','build','-o',str(out/name),'./cmd/asd'],env=env,check=True);hashes[name]=hashlib.sha256((out/name).read_bytes()).hexdigest()
(out/'binary-hashes.json').write_text(json.dumps(hashes,indent=2)+'\n');print(json.dumps(hashes))
