from pathlib import Path
import os,subprocess,json
p=Path('/private/tmp/asd-themes-20261124');env={k:os.environ[k] for k in ['PATH','HOME','TMPDIR','GOCACHE','GOMODCACHE','GOPATH','LANG','TERM'] if k in os.environ};env.update(GOPROXY='off',AGENT_DECK_TEST_HOME_ISOLATED='1',ASD_THEME_REFERENCE_OUT=str(p/'upstream-palette.json'))
for name in ['CONFIG','DATA','CACHE','STATE']:
 directory=p/'reference-xdg'/name.lower();directory.mkdir(parents=True,exist_ok=True);env['XDG_'+name+'_HOME']=str(directory)
config=Path(env['XDG_CONFIG_HOME'])/'agent-deck';config.mkdir(exist_ok=True);(config/'config.toml').write_text('theme = "dark"\n')
command=['sandbox-exec','-f','/private/tmp/asd-upstream-compare-20261123/reference.sb','go','test','-overlay',str(p/'source-overlay.json'),'./internal/ui','-run','^TestASDThemePaletteReference$','-count=1','-v']
with (p/'source-reference.log').open('w') as f:r=subprocess.run(command,cwd='/private/tmp/asd-upstream-compare-20261123/agent-deck',env=env,stdout=f,stderr=subprocess.STDOUT)
assert r.returncode==0,r.returncode
