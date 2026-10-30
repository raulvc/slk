package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/demo"
	"github.com/gammons/slk/internal/ui"
)

// TestDemoNavBackAfterFinderSwitch drives the real demo app through the
// sequence the user reported: switch channels from the finder, then
// alt+shift+left (navigate back). The view must return to #general.
func TestDemoNavBackAfterFinderSwitch(t *testing.T) {
	d, err := demo.New("hero", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	app := newDemoApp(d)
	deliverDemo(t, app, tea.WindowSizeMsg{Width: 140, Height: 40}, 0)
	for _, msg := range d.StartupMsgs() {
		deliverDemo(t, app, msg, 0)
	}
	if app.ActiveChannelID() == "" {
		t.Fatal("no active channel after boot")
	}
	boot := app.ActiveChannelID()

	// Switch to #engineering the way ctrl+t would: openChannel dispatches
	// ChannelSelectedMsg; feed it through Update.
	app.Update(ui.ChannelSelectedMsg{ID: "C0ENGINEERING", Name: "engineering", Type: "channel"})
	if got := app.ActiveChannelID(); got != "C0ENGINEERING" {
		t.Fatalf("active = %q, want C0ENGINEERING after switch", got)
	}

	// alt+shift+left = navigate back.
	_, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt | tea.ModShift})
	if cmd == nil {
		t.Fatal("alt+shift+left dispatched nothing")
	}
	msg := cmd()
	sel, ok := msg.(ui.ChannelSelectedMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ChannelSelectedMsg", msg)
	}
	if sel.ID != boot {
		t.Errorf("back landed on %q, want the boot channel %q", sel.ID, boot)
	}
}
