package main

import (
 "encoding/json"
 "os"
 "github.com/asheshgoplani/agent-deck/internal/ui"
 "github.com/charmbracelet/lipgloss"
 "github.com/muesli/termenv"
)
func main(){
 lipgloss.SetColorProfile(termenv.TrueColor)
 var frames []map[string]any
 for _,size:=range [][2]int{{120,32},{63,25},{49,25},{27,10}} {
  d:=ui.NewGroupDialog();d.SetSize(size[0],size[1]);d.ShowRenameSession("synthetic-id","Copy A")
  frames=append(frames,map[string]any{"width":size[0],"height":size[1],"content":d.View()})
 }
 json.NewEncoder(os.Stdout).Encode(frames)
}
