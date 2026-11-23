package main

import (
 "encoding/json"
 "os"
 "github.com/asheshgoplani/agent-deck/internal/ui"
 "github.com/asheshgoplani/agent-deck/internal/session"
 tea "github.com/charmbracelet/bubbletea"
 "github.com/charmbracelet/lipgloss"
 "github.com/muesli/termenv"
)
func main(){
 lipgloss.SetColorProfile(termenv.TrueColor)
 var frames []map[string]any
 for _,size:=range [][2]int{{120,32},{63,25},{49,25}} {
  s:=ui.NewSearch();s.SetItems([]*session.Instance{{Title:"Copy A",Tool:"shell",ProjectPath:"/synthetic/workspace"},{Title:"Copy B",Tool:"shell",ProjectPath:"/synthetic/workspace"}});s.SetSize(size[0],size[1]);s.Show()
  frames=append(frames,map[string]any{"width":size[0],"height":size[1],"query":"","content":s.View()})
  s.Update(tea.KeyMsg{Type:tea.KeyRunes,Runes:[]rune("Copy B")})
  frames=append(frames,map[string]any{"width":size[0],"height":size[1],"query":"Copy B","content":s.View()})
 }
 json.NewEncoder(os.Stdout).Encode(frames)
}
