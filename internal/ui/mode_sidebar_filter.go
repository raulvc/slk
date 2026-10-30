// internal/ui/mode_sidebar_filter.go
//
// Sidebar-filter mode: `/` while the sidebar (channels bar) has focus.
//
// The prompt lives in the status line's search segment (the same slot
// the in-channel `/` uses — the two never run at once, because they
// are entered from different focus). Every keystroke filters the
// sidebar live via sidebar.SetFilter, which reorders the visible rows
// and resets the cursor to the first match, so the first matching
// channel is highlighted and moved to as the query is typed. j/k (and
// arrows) walk the matches; Enter opens the highlighted channel; Esc
// clears the filter and returns to Normal.
package ui

import (
	tea "charm.land/bubbletea/v2"
)

func handleSidebarFilterMode(a *App, msg tea.KeyMsg) tea.Cmd {
	switch normalizeFinderKey(msg) {
	case "esc":
		a.clearSidebarFilter()
		a.SetMode(ModeNormal)
		return nil
	case "enter":
		// Open the highlighted channel BEFORE clearing: the filter
		// clear resets the sidebar cursor, and handleEnter reads the
		// selection. The ChannelSelectedMsg reducer re-selects the
		// opened channel in the sidebar afterward, so the highlight
		// still ends up on the right row.
		cmd := a.handleEnter()
		a.clearSidebarFilter()
		a.SetMode(ModeNormal)
		return cmd
	case "down":
		a.sidebar.MoveDown()
		return nil
	case "up":
		a.sidebar.MoveUp()
		return nil
	}

	switch s := normalizeFinderKey(msg); s {
	case "backspace":
		if a.sidebarFilterInput != "" {
			r := []rune(a.sidebarFilterInput)
			a.sidebarFilterInput = string(r[:len(r)-1])
		}
	case "space":
		// Key.String() renders a literal space as "space"; channel
		// names can contain spaces (group DMs, user display names).
		a.sidebarFilterInput += " "
	default:
		// Single-rune filter: same known v1 limitation as the
		// in-channel search (drops multi-rune graphemes).
		if len([]rune(s)) == 1 {
			a.sidebarFilterInput += s
		}
	}
	a.sidebar.SetFilter(a.sidebarFilterInput)
	a.statusbar.SetSearch("/" + a.sidebarFilterInput)
	return nil
}

// clearSidebarFilter restores the unfiltered sidebar and the status
// bar's search segment. Shared by the mode's Esc and Enter paths.
func (a *App) clearSidebarFilter() {
	a.sidebarFilterInput = ""
	a.sidebar.SetFilter("")
	a.statusbar.SetSearch("")
}
