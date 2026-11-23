from pathlib import Path
import hashlib,json
p=Path('/private/tmp/asd-key-s-20261123');repo=Path.cwd()
uv=Path('/private/tmp/asd-detach-20261122/go-mod-cache/github.com/charmbracelet/ultraviolet@v0.0.0-20261008173134-6b8d4baf91b4')
reader=uv/'terminal_reader_windows.go';model=repo/'internal/ui/model.go'
s=reader.read_text();a='kevent := record.KeyEvent()';assert s.count(a)==1;s=s.replace(a,a+'\n\t\t\tasdRecordRawKey(kevent)');(p/'terminal_reader_windows.go').write_text(s)
s=model.read_text();a='func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {';assert s.count(a)==1;s=s.replace(a,a+'\n\tm.asdRecordDecodedKey(msg)');(p/'model-debug.go').write_text(s)
common='''import (
 "encoding/json"
 "os"
 "path/filepath"
 "strings"
 "time"
 IMPORTTEA
)
func asdAppendKeyLog(filename string, record map[string]any) {
 root:=os.Getenv("ASD_KEY_DIAG_ROOT");if root=="" {return}
 phase,_:=os.ReadFile(filepath.Join(root,"diag-step.txt"))
 record["tag"]="[DEBUG-keyS]";record["phase"]=strings.TrimSpace(string(phase));record["utc"]=time.Now().UTC().Format(time.RFC3339Nano);record["pid"]=os.Getpid()
 f,err:=os.OpenFile(filepath.Join(root,filename),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{return};defer f.Close();_ =json.NewEncoder(f).Encode(record)
}
'''
raw='//go:build windows\n\npackage uv\n\n'+common.replace(' IMPORTTEA\n','')+'''func asdRecordRawKey(record any) {asdAppendKeyLog("raw-key-events.jsonl",map[string]any{"event":"raw-key","key":record})}
'''
decoded='//go:build windows\n\npackage ui\n\n'+common.replace('IMPORTTEA','tea "charm.land/bubbletea/v2"')+'''func (m *Model) asdRecordDecodedKey(msg tea.Msg) {
 var key tea.Key;kind:=""
 switch v:=msg.(type){case tea.KeyPressMsg:key=v.Key();kind="press";case tea.KeyReleaseMsg:key=v.Key();kind="release";default:return}
 focused:=false;transition:=0
 if m.workspace!=nil {focused=m.workspace.focused;transition=int(m.workspace.transition)}
 asdAppendKeyLog("decoded-key-events.jsonl",map[string]any{"event":kind,"key":key,"string":key.String(),"keystroke":key.Keystroke(),"terminal_focused":focused,"settings_open":m.densitySettings!=nil,"transition":transition})
}
'''
(p/'raw_debug_windows.go').write_text(raw);(p/'decoded_debug_windows.go').write_text(decoded)
overlay={'Replace':{str(reader):str(p/'terminal_reader_windows.go'),str(uv/'asd_debug_windows.go'):str(p/'raw_debug_windows.go'),str(model):str(p/'model-debug.go'),str(repo/'internal/ui/asd_debug_windows.go'):str(p/'decoded_debug_windows.go')}}
(p/'debug-overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
(p/'instrumentation-sources.json').write_text(json.dumps({'scope':'Build-time overlay only; original module and product files remain untouched','files':[{'path':str(f),'sha256':hashlib.sha256(f.read_bytes()).hexdigest()} for f in [reader,model,uv/'decoder.go',uv/'key.go']]},indent=2)+'\n')
