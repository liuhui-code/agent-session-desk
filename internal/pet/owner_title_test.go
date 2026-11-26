package pet

import "testing"

func TestWindowMarkerAcceptsOnlyLauncherAuthoredTokenShape(t *testing.T) {
	if marker, e := WindowMarker("0123456789abcdef0123456789abcdef"); e != nil || marker != "ASD Workbench 0123456789abcdef0123456789abcdef" {
		t.Fatal(marker, e)
	}
	for _, token := range []string{"", "title", "0123456789abcdef0123456789abcde;", "0123456789ABCDEF0123456789ABCDEF"} {
		if _, e := WindowMarker(token); e == nil {
			t.Fatal("invalid window token accepted")
		}
	}
}
