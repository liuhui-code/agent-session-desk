from pathlib import Path
import os,subprocess,json,hashlib
p=Path('/private/tmp/asd-footer-modes-20261124');source=Path('/private/tmp/asd-upstream-compare-20261123/agent-deck')
commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=source,text=True).strip();assert commit=='80494e527393b7ead00da30018e6ee7e60e9d7a0'
assert not subprocess.check_output(['git','status','--porcelain'],cwd=source,text=True)
files=['internal/ui/home_view_frame.go','internal/ui/styles.go','internal/session/userconfig.go','LICENSE'];hashes={n:hashlib.sha256((source/n).read_bytes()).hexdigest() for n in files}
tests=list((source/'internal/ui').glob('*_test.go'));(p/'unused-upstream-tests.go').write_text('package ui\n')
overlay={str(n):str(p/'unused-upstream-tests.go') for n in tests};overlay[str(source/'internal/ui/zz_asd_footer_modes_reference_test.go')]=str(p/'source-reference_test.go');(p/'source-overlay.json').write_text(json.dumps({'Replace':overlay},indent=2)+'\n')
env={k:os.environ[k] for k in ['PATH','HOME','TMPDIR','GOCACHE','GOMODCACHE','GOPATH','LANG','TERM'] if k in os.environ};env.update(GOPROXY='off',AGENT_DECK_TEST_HOME_ISOLATED='1',ASD_FOOTER_REFERENCE_OUT=str(p/'source-reference.json'))
for name in ['CONFIG','DATA','CACHE','STATE']:
 directory=p/'reference-xdg'/name.lower();directory.mkdir(parents=True,exist_ok=True);env['XDG_'+name+'_HOME']=str(directory)
config=Path(env['XDG_CONFIG_HOME'])/'agent-deck';config.mkdir(exist_ok=True);(config/'config.toml').write_text('theme = "dark"\n')
policy=Path('/private/tmp/asd-upstream-compare-20261123/reference.sb')
command=['sandbox-exec','-f',str(policy),'go','test','-overlay',str(p/'source-overlay.json'),'./internal/ui','-run','^TestASDFooterModesReference$','-count=1','-v']
with (p/'source-reference.log').open('w') as f:r=subprocess.run(command,cwd=source,env=env,stdout=f,stderr=subprocess.STDOUT)
assert r.returncode==0,r.returncode
assert hashes=={n:hashlib.sha256((source/n).read_bytes()).hexdigest() for n in files}
status=subprocess.check_output(['git','status','--porcelain'],cwd=source,text=True);assert not status
(p/'upstream-source.json').write_text(json.dumps(dict(commit=commit,worktree=status,source_ui_executed=False,source_helper_executed=True,disabled_ui_tests=len(tests),sha256=hashes),indent=2)+'\n')
(p/'reference-policy.json').write_text(json.dumps(dict(HOME_changed=False,environment_allowlist=True,source_worktree_after='clean',network_outbound_denied=True,real_home_file_access_denied=True,blocked_executables=['security','pbpaste','osascript'],policy_sha256=hashlib.sha256(policy.read_bytes()).hexdigest()),indent=2)+'\n')
