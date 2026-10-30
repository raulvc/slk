package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/ui/sidebar"
	"github.com/gammons/slk/internal/ui/wintree"
)

// sidebarDMOpts is normalOpts with a DM row (DM peer U9) alongside the
// two channels, so the sidebar-P and ctrl+shift+3 rows have a person
// to resolve.
func sidebarDMOpts() []testOpt {
	return append(normalOpts(), withChannels(
		sidebar.ChannelItem{ID: "C1", Name: "general", Type: "channel"},
		sidebar.ChannelItem{ID: "C2", Name: "random", Type: "channel"},
		sidebar.ChannelItem{ID: "D9", Name: "U9", Type: "dm", DMUserID: "U9"},
	))
}

func TestWindowZoomToggle(t *testing.T) {
	a := stackedApp(t, 150)
	if _, err := a.wins.Split(a.focusedWin, wintree.SplitSideBySide, a.windowBounds()); err != nil {
		t.Fatalf("precondition: split failed: %v", err)
	}
	_ = a.View()

	a.toggleWindowZoom()
	// Zoomed: the messages region renders the focused window at full
	// size. The observable: the frame's messages pane uses the full
	// region width instead of a split rect.
	_ = a.View()
	if !a.winZoom {
		t.Fatal("winZoom = false after toggle on")
	}

	a.toggleWindowZoom()
	if a.winZoom {
		t.Fatal("winZoom = true after toggle off")
	}
	_ = a.View()
}

func TestSidebarProfileKey(t *testing.T) {
	runKeyCases(t, ModeNormal, []keyCase{
		{
			name:     "P with a DM row selected opens that person's profile",
			opts:     sidebarDMOpts(),
			setup:    func(t *testing.T, a *App) { a.sidebar.SelectByID("D9") },
			key:      keyPress('P'),
			wantMode: ModeUserProfile,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				if !a.userProfile.IsVisible() {
					t.Fatal("profile dialog not visible")
				}
				if a.mode != ModeUserProfile {
					t.Errorf("mode = %v, want ModeUserProfile", a.mode)
				}
				if cmd == nil {
					t.Fatal("cmd = nil, want the profile fetch cmd")
				}
			},
		},
		{
			name: "P with a channel row selected declines with a toast",
			opts: sidebarDMOpts(),
			setup: func(t *testing.T, a *App) {
				a.sidebar.SelectByID("C1")
			},
			key: keyPress('P'),
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				if a.userProfile.IsVisible() {
					t.Error("profile dialog opened for a channel row")
				}
				if got := statusbarText(a); got == "" {
					t.Error("expected the no-profile toast")
				}
			},
		},
		{
			// The messages-pane path keeps its pre-existing behavior:
			// the author of the selected message.
			name:     "P with the messages pane focused shows the author's profile",
			opts:     normalOpts(),
			setup:    focusMessages,
			key:      keyPress('P'),
			wantMode: ModeUserProfile,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				if !a.userProfile.IsVisible() {
					t.Fatal("profile dialog not visible")
				}
			},
		},
	})
}

func TestJumpToDMs(t *testing.T) {
	// One collapsed section (Channels) holding both channels, and the
	// default-fallback DM section holding D9: the jump must expand and
	// land on D9.
	a := newTestApp(t, sidebarDMOpts()...)
	a.Update(tea.KeyPressMsg{Code: '3', Mod: tea.ModCtrl | tea.ModShift})

	if a.focusedPanel != PanelSidebar {
		t.Errorf("focusedPanel = %v, want PanelSidebar", a.focusedPanel)
	}
	if got := a.sidebar.SelectedID(); got != "D9" {
		t.Errorf("sidebar SelectedID = %q, want the first DM D9", got)
	}
}

func TestThreadsViewToggleKey(t *testing.T) {
	a := newTestApp(t, normalOpts()...)
	// Enter the threads view through the key.
	_, cmd := a.Update(keyMod('t', tea.ModCtrl|tea.ModShift))
	if cmd == nil {
		t.Fatal("ctrl+shift+t dispatched nothing")
	}
	if _, ok := cmd().(ThreadsViewActivatedMsg); !ok {
		t.Fatalf("cmd() = %T, want ThreadsViewActivatedMsg", cmd())
	}
	// Feed the activation through Update (bubbletea would) so the
	// view actually switches; the activation's own fetch cmds are
	// dropped, they only populate the list.
	a.Update(cmd())

	// Pressing it again returns to the channel view.
	_, cmd = a.Update(keyMod('t', tea.ModCtrl|tea.ModShift))
	if cmd != nil {
		t.Fatalf("toggle-off dispatched %T, want nil", cmd())
	}
	if a.view != ViewChannels {
		t.Errorf("view = %v, want ViewChannels after toggle-off", a.view)
	}
}
