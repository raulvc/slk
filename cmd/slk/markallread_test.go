package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/cache"
	"github.com/gammons/slk/internal/ui"
)

// seedMarkAllDB builds a T1 workspace with:
//
//	C1  unread, watermark from a cached message
//	C2  unread (with a mention badge), watermark from latest_synced_ts
//	C3  read (never a mark target)
//	C4  unread but never synced: no messages, no latest_synced_ts
func seedMarkAllDB(t *testing.T, db *cache.DB) {
	t.Helper()
	channels := []struct {
		id, name string
	}{
		{"C1", "general"},
		{"C2", "random"},
		{"C3", "quiet"},
		{"C4", "never-synced"},
	}
	for _, ch := range channels {
		if err := db.UpsertChannel(cache.Channel{ID: ch.id, WorkspaceID: "T1", Name: ch.name, Type: "channel"}); err != nil {
			t.Fatalf("UpsertChannel %s: %v", ch.id, err)
		}
	}
	states := []struct {
		id        string
		lastRead  string
		hasUnread bool
	}{
		{"C1", "1700000000.000050", true},
		{"C2", "", true},
		{"C3", "1700000000.000010", false},
		{"C4", "1700000000.000020", true},
	}
	for _, s := range states {
		if err := db.UpdateChannelReadState(s.id, s.lastRead, s.hasUnread); err != nil {
			t.Fatalf("UpdateChannelReadState %s: %v", s.id, err)
		}
	}
	if err := db.SetChannelMentionCount("C2", 3); err != nil {
		t.Fatalf("SetChannelMentionCount: %v", err)
	}
	if err := db.UpsertMessage(cache.Message{
		ChannelID:   "C1",
		WorkspaceID: "T1",
		TS:          "1700000000.000100",
		UserID:      "U2",
		Text:        "newest message in C1",
	}); err != nil {
		t.Fatalf("UpsertMessage: %v", err)
	}
	if err := db.SetChannelLatestSyncedTS("C2", "1700000001.000000"); err != nil {
		t.Fatalf("SetChannelLatestSyncedTS: %v", err)
	}
	if err := db.SetChannelLatestSyncedTS("C3", "1700000000.000010"); err != nil {
		t.Fatalf("SetChannelLatestSyncedTS: %v", err)
	}
}

func TestMarkAllChannelsRead_MarksUnreadOnly(t *testing.T) {
	db := newTestDB(t)
	seedMarkAllDB(t, db)
	marker := &fakeChannelMarker{}
	var sent []tea.Msg

	n := markAllChannelsRead(context.Background(), marker, db, func(m tea.Msg) {
		sent = append(sent, m)
	}, "T1")

	want := []string{"C1/1700000000.000100", "C2/1700000001.000000"}
	slices.Sort(marker.calls)
	if n != 2 || !slices.Equal(marker.calls, want) {
		t.Fatalf("marked=%d calls=%v, want %v", n, marker.calls, want)
	}
	if len(sent) != 2 {
		t.Fatalf("notify delivered %d messages, want 2: %v", len(sent), sent)
	}
	for _, m := range sent {
		msg, ok := m.(ui.ChannelMarkedReadMsg)
		if !ok {
			t.Fatalf("notify delivered %T, want ui.ChannelMarkedReadMsg", m)
		}
		if msg.ChannelID != "C1" && msg.ChannelID != "C2" {
			t.Errorf("notify delivered a mark for %q, want C1 or C2", msg.ChannelID)
		}
	}
	for _, ch := range []string{"C1", "C2"} {
		s, err := db.GetChannelReadState(ch)
		if err != nil {
			t.Fatalf("GetChannelReadState %s: %v", ch, err)
		}
		if s.HasUnread {
			t.Errorf("C%s HasUnread = true, want false", ch)
		}
	}
	if s, _ := db.GetChannelReadState("C2"); s.MentionCount != 0 {
		t.Errorf("C2 MentionCount = %d, want 0 after mark", s.MentionCount)
	}
	if s, _ := db.GetChannelReadState("C3"); s.HasUnread {
		t.Error("C3 (already read) must not be marked")
	}
}

func TestMarkAllChannelsRead_SkipsChannelWithoutWatermark(t *testing.T) {
	db := newTestDB(t)
	seedMarkAllDB(t, db)
	marker := &fakeChannelMarker{}

	n := markAllChannelsRead(context.Background(), marker, db, nil, "T1")

	for _, call := range marker.calls {
		if strings.HasPrefix(call, "C4/") {
			t.Errorf("C4 has no watermark and must be skipped, got call %q", call)
		}
	}
	if n != 2 {
		t.Fatalf("marked=%d, want 2 (C1 and C2; C4 skipped)", n)
	}
	if s, _ := db.GetChannelReadState("C4"); !s.HasUnread {
		t.Error("C4 must stay unread so channel entry can mark it")
	}
}

func TestMarkAllChannelsRead_RejectedMarkDoesNotStopWalk(t *testing.T) {
	db := newTestDB(t)
	seedMarkAllDB(t, db)
	marker := &fakeChannelMarker{err: errors.New("conversations.mark: rate_limited")}
	var sent []tea.Msg

	n := markAllChannelsRead(context.Background(), marker, db, func(m tea.Msg) {
		sent = append(sent, m)
	}, "T1")

	if n != 2 {
		t.Fatalf("attempts=%d, want 2 (the walk continues past a rejection)", n)
	}
	if len(marker.calls) != 2 {
		t.Fatalf("attempts=%v, want 2 (the walk continues past a rejection)", marker.calls)
	}
	if len(sent) != 0 {
		t.Errorf("notify called with %v, want nothing for rejected marks", sent)
	}
}

func TestMarkAllChannelsRead_UnknownWorkspaceIsInert(t *testing.T) {
	db := newTestDB(t)
	seedMarkAllDB(t, db)
	marker := &fakeChannelMarker{}

	if n := markAllChannelsRead(context.Background(), marker, db, nil, "T2"); n != 0 {
		t.Fatalf("marked=%d, want 0 for a workspace with no read state", n)
	}
	if len(marker.calls) != 0 {
		t.Fatalf("calls=%v, want none", marker.calls)
	}
}

func TestMarkAllChannelsReadAsync_GuardsBeforeGoroutine(t *testing.T) {
	// The guards must return before the goroutine starts, so a nil
	// client (workspace not connected) and a nil db cannot panic inside
	// it. The sync core's own tests cover the marked path; here only
	// the pre-goroutine contract matters.
	markAllChannelsReadAsync(context.Background(), nil, nil, nil, "T1")

	db := newTestDB(t)
	seedMarkAllDB(t, db)
	markAllChannelsReadAsync(context.Background(), nil, db, nil, "T1")
	markAllChannelsReadAsync(context.Background(), &fakeChannelMarker{}, nil, nil, "T1")
}
