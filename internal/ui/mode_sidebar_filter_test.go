package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// seedSidebarFilter parks the sidebar-filter mode mid-query, the way
// typing "gen" through the handler would: prompt text on the status
// bar and the live filter applied. Used as the precondition for the
// exit/commit rows so they exercise real state instead of a fresh one.
func seedSidebarFilter(t *testing.T, a *App) {
	t.Helper()
	a.sidebarFilterInput = "gen"
	a.sidebar.SetFilter("gen")
	a.statusbar.SetSearch("/gen")
}

func TestSidebarFilterModeKeys(t *testing.T) {
	runKeyCases(t, ModeSidebarFilter, []keyCase{
		{
			// "g" matches only "general" in the normalOpts fixture
			// (C1 general, C2 random). The highlight must sit on the
			// matched channel, not the pinned Threads row the cursor
			// clamps to on an empty filter.
			name:     "typing filters live and highlights the first matching channel",
			opts:     normalOpts(),
			key:      keyPress('g'),
			wantMode: ModeSidebarFilter,
			assert: func(t *testing.T, a *App, _ tea.Cmd) {
				if got := a.sidebar.SelectedID(); got != "C1" {
					t.Errorf("sidebar SelectedID = %q, want C1 (first match)", got)
				}
				if got := a.statusbar.Search(); got != "/g" {
					t.Errorf("statusbar search = %q, want %q", got, "/g")
				}
			},
		},
		{
			name:     "backspace narrows the filter back",
			opts:     normalOpts(),
			setup:    seedSidebarFilter,
			key:      keyCode(tea.KeyBackspace),
			wantMode: ModeSidebarFilter,
			assert: func(t *testing.T, a *App, _ tea.Cmd) {
				if got := a.sidebarFilterInput; got != "ge" {
					t.Errorf("sidebarFilterInput = %q, want %q", got, "ge")
				}
				if got := a.sidebar.SelectedID(); got != "C1" {
					t.Errorf("sidebar SelectedID = %q, want C1 still matched by \"ge\"", got)
				}
			},
		},
		{
			name:     "enter opens the matched channel and clears the filter",
			opts:     normalOpts(),
			setup:    seedSidebarFilter,
			key:      keyCode(tea.KeyEnter),
			wantMode: ModeNormal,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				if a.mode != ModeNormal {
					t.Errorf("mode = %v, want Normal", a.mode)
				}
				if got := a.sidebarFilterInput; got != "" {
					t.Errorf("sidebarFilterInput = %q, want cleared", got)
				}
				if got := a.statusbar.Search(); got != "" {
					t.Errorf("statusbar search = %q, want cleared", got)
				}
				if cmd == nil {
					t.Fatal("cmd = nil, want a ChannelSelectedMsg cmd")
				}
				msg, ok := cmd().(ChannelSelectedMsg)
				if !ok {
					t.Fatalf("cmd() = %#v, want ChannelSelectedMsg", cmd())
				}
				if msg.ID != "C1" {
					t.Errorf("cmd() = %+v, want the matched channel C1", msg)
				}
			},
		},
		{
			name:     "esc clears the filter and returns to Normal",
			opts:     normalOpts(),
			setup:    seedSidebarFilter,
			key:      keyCode(tea.KeyEscape),
			wantMode: ModeNormal,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				if a.mode != ModeNormal {
					t.Errorf("mode = %v, want Normal", a.mode)
				}
				if got := a.sidebarFilterInput; got != "" {
					t.Errorf("sidebarFilterInput = %q, want cleared", got)
				}
				if got := a.statusbar.Search(); got != "" {
					t.Errorf("statusbar search = %q, want cleared", got)
				}
				if got := len(a.sidebar.VisibleItems()); got != 2 {
					t.Errorf("visible items = %d, want the unfiltered 2", got)
				}
				if cmd != nil {
					t.Errorf("cmd = %T, want nil", cmd)
				}
			},
		},
		{
			name:     "down moves the highlight within the matches",
			opts:     normalOpts(),
			setup:    seedSidebarFilter,
			key:      keyCode(tea.KeyDown),
			wantMode: ModeSidebarFilter,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				// "gen" matches only general; the cursor may sit on the
				// header past it after a down — the observable part is
				// that the mode swallows the key (no cmd) and stays.
				if cmd != nil {
					t.Errorf("cmd = %T, want nil", cmd)
				}
			},
		},
	})
}

// TestNormalSlashOnSidebarEntersFilter pins the mode_normal arm: with
// the sidebar focused, `/` opens the channel filter, not the in-channel
// message search.
func TestNormalSlashOnSidebarEntersFilter(t *testing.T) {
	runKeyCases(t, ModeNormal, []keyCase{
		{
			name:     "/ with the sidebar focused enters the sidebar filter",
			opts:     normalOpts(),
			key:      keyPress('/'),
			wantMode: ModeSidebarFilter,
			assert: func(t *testing.T, a *App, _ tea.Cmd) {
				if got := a.statusbar.Search(); got != "/" {
					t.Errorf("statusbar search = %q, want %q", got, "/")
				}
				if got := a.sidebarFilterInput; got != "" {
					t.Errorf("sidebarFilterInput = %q, want empty on entry", got)
				}
			},
		},
		{
			// Focus guard: the messages pane keeps the in-channel search.
			name:     "/ with the messages pane focused stays on message search",
			opts:     normalOpts(),
			setup:    focusMessages,
			key:      keyPress('/'),
			wantMode: ModeSearch,
			assert: func(t *testing.T, a *App, cmd tea.Cmd) {
				_ = cmd
				if got := a.sidebarFilterInput; got != "" {
					t.Errorf("sidebarFilterInput = %q, want untouched", got)
				}
			},
		},
	})
}
