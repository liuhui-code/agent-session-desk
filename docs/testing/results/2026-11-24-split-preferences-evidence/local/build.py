from pathlib import Path
import subprocess,os,json,hashlib
out=Path('/private/tmp/asd-split-preferences-20261124');env=dict(os.environ,GOOS='windows',GOARCH='amd64',CGO_ENABLED='0');hashes={}
for package in ['config','cli','ui']:
 name='split-preferences-'+package+'.exe';subprocess.run(['go','test','-c','-o',str(out/name),'./internal/'+package],env=env,check=True);hashes[name]=hashlib.sha256((out/name).read_bytes()).hexdigest()
name='asd-split-preferences.exe';subprocess.run(['go','build','-o',str(out/name),'./cmd/asd'],env=env,check=True);hashes[name]=hashlib.sha256((out/name).read_bytes()).hexdigest()
(out/'binary-hashes.json').write_text(json.dumps(hashes,indent=2)+'\n');print(json.dumps(hashes))
