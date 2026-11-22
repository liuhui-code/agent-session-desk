import time,json
from ref_control import tmux,ROOT
from capture_ref import capture
log=[]
def send(*keys,literal=False):
 args=['send-keys','-t','reference:0.0']+(['-l'] if literal else [])+list(keys)
 r=tmux(args);log.append({'utc':time.time(),'args':args,'code':r.returncode});assert r.returncode==0,r.stderr
 (ROOT/'reference-actions.json').write_text(json.dumps(log,indent=2)+'\n')
def size(cols,rows):
 r=tmux(['resize-window','-t','reference:0','-x',str(cols),'-y',str(rows)]);assert r.returncode==0;log.append({'utc':time.time(),'resize':[cols,rows]});time.sleep(.3)
# The selected session and flat mode were reached through actual upstream keys.
capture('upstream-manager-120-clean')
send('S');time.sleep(.3);capture('upstream-settings-120')
send('Escape');time.sleep(.2)
for seq in ['\x1b[<0;22;10M','\x1b[<32;43;10M','\x1b[<0;43;10m']:
 send(seq,literal=True);time.sleep(.1)
time.sleep(1.7);capture('upstream-manager-wide-120')
send('Enter');time.sleep(.7);capture('upstream-focused-wide-120')
send('C-q');time.sleep(5.5)
size(63,25);capture('upstream-manager-63')
send('Enter');time.sleep(.6);capture('upstream-focused-63')
send('C-q');time.sleep(5.5)
size(49,25);capture('upstream-manager-49')
send('S');time.sleep(.3);capture('upstream-settings-49')
send('Escape');size(120,32)
(ROOT/'reference-actions.json').write_text(json.dumps(log,indent=2)+'\n')
