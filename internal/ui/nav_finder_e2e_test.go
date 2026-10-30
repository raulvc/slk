package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/core"
	"github.com/gammons/slk/internal/ids"
)

// TestFinderSwitchThenNavBack drives the real production sequence:
// open a channel from the finder (ctrl+t path), then navigate back with
// ctrl+h. The nav history must contain both channels and back must
// return to the first one.
func TestFinderSwitchThenNavBack(t *testing.T) {
	lookupCalls := 0
	a := newTestApp(t,
		append(normalOpts(), withActiveTeam("T1"), withChannelService(core.ChannelServiceFuncs{
			Lookup: func(id ids.ChannelID) (string, string, bool) {
				lookupCalls++
				switch string(id) {
				case "C1":
					return "general", "channel", true
				case "C2":
					return "random", "channel", true
				}
				return "", "", false
			},
		}))...,
	)

	// Boot selects the starting channel through the real reducer path
	// (reduceWorkspaceReady dispatches ChannelSelectedMsg), which is
	// what pushes it onto the nav history.
	a.Update(ChannelSelectedMsg{ID: "C1", Name: "general", Type: "channel"})

	// The finder path: ctrl+t opens the finder, enter on C2 selects it.
	_, _ = a.Update(keyMod('t', tea.ModCtrl))
	a.channelFinder.SetItems([]core.ChannelFinderItem{
		{ID: "C1", Name: "general", Type: "channel", Joined: true},
		{ID: "C2", Name: "random", Type: "channel", Joined: true},
	})
	for _, r := range "rand" {
		a.channelFinder.HandleKey(string(r))
	}
	res := a.channelFinder.HandleKey("enter")
	if res == nil {
		t.Fatal("finder enter returned nil")
	}
	// Route the result the way handleChannelFinderMode does.
	a.channelFinder.Close()
	a.SetMode(ModeNormal)
	if cmd := a.openChannel(*res); cmd != nil {
		a.Update(cmd())
	}

	if a.activeChannelID != "C2" {
		t.Fatalf("activeChannelID = %q, want C2 after finder switch", a.activeChannelID)
	}
	stack := a.navHistory.Stack("T1")
	if stack == nil {
		t.Fatal("nav stack not created after finder switch")
	}
	t.Logf("stack: cursor=%d entries=%+v", stack.cursor, stack.entries)

	// Navigate back.
	_, cmd := a.Update(keyMod('h', tea.ModCtrl))
	if cmd == nil {
		t.Fatal("ctrl+h dispatched nothing")
	}
	msg := cmd()
	sel, ok := msg.(ChannelSelectedMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ChannelSelectedMsg", msg)
	}
	if sel.ID != "C1" || !sel.FromHistory {
		t.Errorf("cmd() = %+v, want {C1 FromHistory:true}", sel)
	}
}
