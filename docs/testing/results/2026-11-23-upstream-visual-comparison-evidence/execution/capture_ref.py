from pathlib import Path
import json,time,sys,hashlib,re
from ref_control import tmux,ROOT
FRAMES=ROOT/'frames';FRAMES.mkdir(exist_ok=True)
def capture(name):
 r=tmux(['display-message','-p','-t','reference:0.0','#{pid}|#{pane_pid}|#{pane_width}|#{pane_height}|#{pane_dead}|#{cursor_flag}|#{cursor_x}|#{cursor_y}'])
 assert r.returncode==0,r.stderr
 fields=r.stdout.decode().strip().split('|');server,pid,width,height,dead,visible,x,y=map(int,fields)
 assert dead==0,fields
 raw=tmux(['capture-pane','-p','-e','-t','reference:0.0']).stdout
 text=tmux(['capture-pane','-p','-t','reference:0.0']).stdout
 assert b'<LOCAL_TEST_HOME>' not in raw,'Private path present in reference frame'
 rows=raw.split(b'\n');assert rows[-1]==b'';rows.pop();assert len(rows)==height,(len(rows),height)
 vt=b'\x1b[0m\x1b[?25l\x1b[2J\x1b[H'+b'\r\n'.join(rows)+b'\x1b[0m'
 if visible:vt+=f'\x1b[{y+1};{x+1}H\x1b[?25h'.encode()
 (FRAMES/(name+'.capture.ansi')).write_bytes(raw);(FRAMES/(name+'.txt')).write_bytes(text);(FRAMES/(name+'.vt')).write_bytes(vt)
 manifest=FRAMES/'manifest.json';data=json.loads(manifest.read_text()) if manifest.exists() else {}
 data[name]={'file':name+'.vt','width':width,'height':height,'cursor_visible':bool(visible),'cursor_x':x,'cursor_y':y,'server_pid':server,'pane_pid':pid,'capture_sha256':hashlib.sha256(raw).hexdigest(),'vt_sha256':hashlib.sha256(vt).hexdigest()}
 manifest.write_text(json.dumps(data,indent=2)+'\n');print(name,width,height,'cursor',visible,x,y);print(text.decode())
if __name__=='__main__':capture(sys.argv[1])
