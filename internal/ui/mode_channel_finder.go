// internal/ui/mode_channel_finder.go
//
// Channel-finder mode key handler (Phase 5h).
//
// Forwards normalised keys to the channel-finder overlay. On a
// result:
//   - Synthetic "threads"/"activity" destinations -> activate the
//     corresponding view (ThreadsViewActivatedMsg /
//     ActivityViewActivatedMsg).
//   - Already-joined channel -> select it (ChannelSelectedMsg).
//   - Not yet joined -> fire a Join via the channels service;
//     ChannelJoinedMsg (reducer_channels) folds it into the
//     sidebar and dispatches the ChannelSelectedMsg.
package ui

import (
	tea "charm.land/bubbletea/v2"
)

func handleChannelFinderMode(a *App, msg tea.KeyMsg) tea.Cmd {
	// Map tea.KeyMsg to string for the finder.
	before := a.channelFinder.Query()
	result := a.channelFinder.HandleKey(normalizeFinderKey(msg))
	if result != nil {
		if a.pendingForward != nil {
			return a.forwardToChannel(*result)
		}
		a.channelFinder.Close()
		a.SetMode(ModeNormal)
		// Synthetic destinations (e.g. Threads, Activity) live alongside
		// channels in the finder but route to a view activation rather
		// than a channel switch.
		switch result.Type {
		case "threads":
			return func() tea.Msg { return ThreadsViewActivatedMsg{} }
		case "activity":
			// Same precondition as the ctrl+a path: the sidebar row is
			// marked active before the view activates.
			a.sidebar.SelectActivityRow()
			return func() tea.Msg { return ActivityViewActivatedMsg{} }
		}
		return a.openChannel(*result)
	}

	// Check if finder closed itself (Esc).
	if !a.channelFinder.IsVisible() {
		a.SetMode(ModeNormal)
		return nil
	}

	// The local filter has already run inside HandleKey, so the list
	// on screen is up to date before anything touches the network.
	// Only the server query is deferred.
	if after := a.channelFinder.Query(); after != before && a.pendingForward == nil {
		return a.scheduleChannelSearch(after)
	}
	return nil
}
