from pathlib import Path,PurePosixPath
import hashlib,json,os,subprocess,zipfile
scratch=Path('/private/tmp/asd-custom-groups-20261123')
root=Path('docs/testing/results/2026-11-23-custom-groups-evidence')
root.mkdir(parents=True,exist_ok=True)
entries=[]
def sha(b):return hashlib.sha256(b).hexdigest()
def put(dest,raw,origin):
    if raw[:2] in [b'\xff\xfe',b'\xfe\xff']:text=raw.decode('utf-16')
    else:text=raw.decode('utf-8-sig')
    text=text.replace('\r\n','\n').replace('\r','\n').replace(os.environ.get('ASD_EVIDENCE_WINDOWS_USER','<WINDOWS_TEST_USER>'),'<WINDOWS_TEST_USER>').replace(str(Path.cwd()),'<REPO>').replace(str(Path.home()),'<LOCAL_TEST_HOME>')
    data=text.encode('utf-8');p=root/dest;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(data)
    entries.append({'path':dest,'origin':origin,'source_sha256':sha(raw),'sha256':sha(data),'bytes':len(data)})
for remote,local in [('custom-groups-model-run3-export.zip','native-model-final.zip')]:
 subprocess.run(['scp','windows-pc:asd-acceptance-20261122/'+remote,str(scratch/local)],check=True,capture_output=True)
expected={'native-model-before-placeholders.zip':'5e59c4eb9409085d89a2f3ebdf956e11109f837d015de730d260e15cbfe34c85','native-model-final.zip':'14c377af44431b4d58168bd83cdc2b1000a803e24c3e9213882dbce2749ff6b9'}
for zipname,label in [('native-model-before-placeholders.zip','first'),('native-model-final.zip','final')]:
 path=scratch/zipname
 assert sha(path.read_bytes())==expected[zipname],('ZIP checksum',zipname)
 with zipfile.ZipFile(path) as z:
  for info in z.infolist():
   if info.is_dir():continue
   name=PurePosixPath(info.filename.replace('\\','/')).name
   assert name in {'results.json','environment.json','session-tests.log','store-tests.log','manager-tests.log','config-tests.log','ui-tests.log'},name
   put('native/'+label+'/'+name,z.read(info),zipname+':'+info.filename)
for name in ['store-tests.log','session-tests.log','results.json']:
 p=scratch/('native-final-first-'+name)
 put('native/final-first-failed/'+name,p.read_bytes(),'Windows failed final run')
for scope,remote in [('observer-red','custom-groups-observer-red-20261123'),('observer-green','custom-groups-observer-green-20261123')]:
 for name in [scope+'.log','result.json']:
  p=scratch/(scope+'-'+name)
  subprocess.run(['scp','windows-pc:asd-acceptance-20261122/'+remote+'/'+name,str(p)],check=True,capture_output=True)
  put('native/'+scope+'/'+name,p.read_bytes(),remote+'/'+name)
# Explicit extensions, excluding raw downloaded native entries already handled.
for p in sorted(scratch.iterdir()):
 if not p.is_file() or p.suffix not in {'.log','.json','.ps1','.py'}:continue
 if p.name.startswith(('native-final-first-','observer-red-result','observer-green-result','observer-red-observer','observer-green-observer')):continue
 put('local/'+p.name,p.read_bytes(),'local validation workspace')
# Verify source snapshots still match the final built/tested tree.
source=json.loads((scratch/'source-before.json').read_text())
changed=[v['path'] for v in source['files'] if sha(Path(v['path']).read_bytes())!=v['sha256']]
assert not changed,changed
put('local/source-after.json',(json.dumps(source,indent=2)+'\n').encode(),'current worktree, same hashes as final build')
upstream=Path('/private/tmp/asd-upstream-compare-20261123/agent-deck')
u=json.loads((scratch/'upstream-source.json').read_text())
assert subprocess.check_output(['git','status','--porcelain'],cwd=upstream,text=True)==''
for v in u['files']:assert sha((upstream/v['path']).read_bytes())==v['sha256'],v['path']
manifest={'schema_version':1,'date':'2026-11-23','base_commit':source['base_commit'],'scope':'persistent group create/move, local and native model/store/config; no fresh desktop claim','zip_sha256':expected,'redaction':['Windows synthetic test user','local repository/home paths'],'files':sorted(entries,key=lambda e:e['path'])}
(root/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
(root/'README.md').write_text('# 分组局部验证证据\n\n`manifest.json` 列出全部允许归档文件、原始字节与归档 UTF-8/LF 哈希；不含二进制、凭据、活目录或用户会话。`native/first` 为占位提示前首轮，`native/final-first-failed` 保留共享占用失败，`native/observer-red`／`observer-green` 为确定性探针与各 20 次复验，`native/final` 为最终五套件。`local` 保留 TDD、先后源码指纹、检查和复现脚本。\n\n原生模型不是桌面证据；当前产品仅本地构建并记哈希，未在本轮启动真实终端。源码指纹含 Go 代码与测试、go.mod/go.sum；170→172 是新增平台观察器测试文件，产品运行代码未因此改变。原失败不覆盖，339 项设计状态不提升。\n')
print(json.dumps({'files':len(entries),'source_files':len(source['files']),'changed_sources':changed,'zips_verified':len(expected)}))
