from pathlib import Path,PurePosixPath
import hashlib,json,os,zipfile
p=Path('/private/tmp/asd-groups-desktop-20261123');root=Path('docs/testing/results/2026-11-23-groups-desktop-evidence');root.mkdir(parents=True,exist_ok=True)
sha=lambda b:hashlib.sha256(b).hexdigest()
win_user=os.environ['ASD_EVIDENCE_WINDOWS_USER'];entries=[]
def add(dest,raw,origin,binary=False):
 if binary:data=raw
 else:
  text=raw.decode('utf-16')if raw[:2]in[b'\xff\xfe',b'\xfe\xff']else raw.decode('utf-8-sig')
  text=text.replace('\r\n','\n').replace('\r','\n').replace(win_user,'<WINDOWS_TEST_USER>').replace(str(Path.cwd()),'<REPO>').replace(str(Path.home()),'<LOCAL_TEST_HOME>')
  data=text.encode('utf-8')
 target=root/dest;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
 entries.append({'path':dest,'origin':origin,'source_sha256':sha(raw),'sha256':sha(data),'bytes':len(data),'binary_unchanged':binary})
zip_sha='b0376b122c72f35c0771fc43ec6573368c63d8c788943d0f2a8edf55a969f934'
assert sha((p/'desktop-export.zip').read_bytes())==zip_sha
with zipfile.ZipFile(p/'desktop-export.zip')as z:
 for e in z.infolist():
  rel=PurePosixPath(e.filename.replace('\\','/'));assert not rel.is_absolute()and'..'not in rel.parts
  if e.is_dir():continue
  assert rel.suffix in {'.json','.log','.png','.toml','.ps1'},str(rel)
  assert not any(part in {'.key','.port','runtime','state'}for part in rel.parts),str(rel)
  add('native/'+str(rel),z.read(e),'desktop-export.zip:'+e.filename,rel.suffix=='.png')
for f in sorted(p.iterdir()):
 if not f.is_file()or f.suffix not in {'.json','.log','.ps1','.py','.go','.txt','.sb'}:continue
 name=f.name+'.txt'if f.suffix=='.go'else f.name
 add('local/'+name,f.read_bytes(),'local validation workspace')
source=json.loads((p/'product-source-before.json').read_text())
assert all(sha(Path(f['path']).read_bytes())==f['sha256']for f in source['files'])
add('local/product-source-after.json',(json.dumps(source,indent=2)+'\n').encode(),'current worktree, identical source hashes')
(root/'README.md').write_text('# 分组来源与桌面证据\n\n`manifest.json` 是全部允许归档文件的清单，记录原字节及脱敏 UTF-8/LF 哈希，PNG 原字节不变。`native` 为完整窗口／观察／清理白名单，`local` 为 RED/GREEN、实际来源／产品帧、helper、固定源码、构建及执行记录。Go helper 以 `.go.txt` 保存，避免成为额外 Go 包。\n\n不包含数据库、auth key/port、设置备份、用户或测试历史内容、原剪贴板数据、测试二进制；第086捕获因程序字段显示测试用户路径而排除图片，仅保留脱敏回执及原图哈希。源码／模块174项指纹在构建、原生执行和归档后相同。脚本的个人目录／仓库路径已脱敏，复现时须映射相应的隔离环境。\n')
x={'schema_version':1,'date':'2026-11-23','base_commit':source['commit'],'scope':'group component references, CJK placeholder fix and bounded actual Windows three-shell desktop flow','desktop_zip_sha256':zip_sha,'omitted_original_images':[{'path':'desktop/new-in-child-default-directory-public.png','source_sha256':sha((p/'images/new-in-child-default-directory-public.png').read_bytes()),'reason':'test user path visible; original image not edited or committed'}],'files':sorted(entries,key=lambda e:e['path'])}
(root/'manifest.json').write_text(json.dumps(x,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'artifacts':len(entries),'native_artifacts':len([e for e in entries if e['path'].startswith('native/')]),'png':len([e for e in entries if e['binary_unchanged']]),'source_files':len(source['files'])}))
