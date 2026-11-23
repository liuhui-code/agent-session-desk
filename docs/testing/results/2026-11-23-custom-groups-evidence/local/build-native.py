from pathlib import Path
import hashlib,json,os,subprocess
out=Path('/private/tmp/asd-custom-groups-20261123')
files=sorted([*Path('internal').rglob('*.go'),*Path('cmd').rglob('*.go'),Path('go.mod'),Path('go.sum')])
snapshot={'base_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'go_version':subprocess.check_output(['go','version'],text=True).strip(),'files':[{'path':str(p),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}for p in files]}
(out/'source-before.json').write_text(json.dumps(snapshot,indent=2)+'\n')
env={**os.environ,'GOOS':'windows','GOARCH':'amd64','CGO_ENABLED':'0'}
binaries=[]
for package in ['session','store','manager','config','ui']:
    path=out/('groups-'+package+'-tests.exe')
    command=['go','test','-c','-o',str(path),'./internal/'+package]
    result=subprocess.run(command,env=env,capture_output=True)
    (out/('build-'+package+'.log')).write_bytes(result.stdout+result.stderr)
    if result.returncode:raise SystemExit(result.returncode)
    binaries.append({'package':package,'file':path.name,'sha256':hashlib.sha256(path.read_bytes()).hexdigest()})
path=out/'asd-custom-groups.exe'
result=subprocess.run(['go','build','-o',str(path),'./cmd/asd'],env=env,capture_output=True)
(out/'build-product.log').write_bytes(result.stdout+result.stderr)
if result.returncode:raise SystemExit(result.returncode)
binaries.append({'package':'product','file':path.name,'sha256':hashlib.sha256(path.read_bytes()).hexdigest()})
(out/'binary-hashes.json').write_text(json.dumps(binaries,indent=2)+'\n')
print(json.dumps({'binaries':binaries,'source_files':len(files)},indent=2))
