from pathlib import Path
import json,os,subprocess,sys,time,shlex,hashlib
ROOT=Path('/private/tmp/asd-upstream-compare-20261123')
DATA=Path('/private/tmp/asdui-1123')
APP=ROOT/'agent-deck-darwin'
OUTER='asdref-outer-1123';INNER='asdref-inner-1123'
def env():
 e={key:os.environ[key] for key in ['HOME','USER','LOGNAME','LANG'] if key in os.environ}
 e.update(PATH=str(DATA/'bin')+':/usr/bin:/bin:/usr/sbin:/sbin',XDG_CONFIG_HOME=str(DATA/'c'),XDG_DATA_HOME=str(DATA/'d'),XDG_CACHE_HOME=str(DATA/'k'),XDG_STATE_HOME=str(DATA/'s'),TMUX_TMPDIR=str(DATA/'t'),TMPDIR=str(DATA/'tmp'),TERM='xterm-256color',COLORTERM='truecolor',AGENTDECK_COLOR='truecolor',AGENTDECK_PROFILE='compare',AGENTDECK_TELEMETRY='0',AGENTDECK_SKIP_UPDATE_CHECK='1',DO_NOT_TRACK='1')
 return e
def command(args,timeout=45):
 return subprocess.run(['/usr/bin/sandbox-exec','-f',str(ROOT/'reference.sb'),*map(str,args)],cwd=DATA/'work',env=env(),capture_output=True,timeout=timeout)
def cli(args):return command([APP,'-p','compare',*args])
def tmux(args,inner=False):return command(['/usr/local/bin/tmux','-f','/dev/null','-L',INNER if inner else OUTER,*args])
if __name__=='__main__':
 mode=sys.argv[1];args=sys.argv[2:]
 r=cli(args) if mode=='cli' else tmux(args,mode=='inner')
 sys.stdout.buffer.write(r.stdout);sys.stderr.buffer.write(r.stderr);sys.exit(r.returncode)
