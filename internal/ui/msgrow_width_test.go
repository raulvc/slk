package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/gammons/slk/internal/emoji"
	"github.com/gammons/slk/internal/ui/messages"
)

// TestMessagesPaneRowsExactWidth pins the border-assembly width
// invariant: every line ViewBare emits must measure exactly the pane
// content width, because renderMessagesTop concatenates the side
// borders around them without re-measuring. A short line leaves the
// border one (or more) cells to the LEFT of the pane's other rows —
// the misaligned-border artifact; a long one pushes it out.
//
// Run with image mode both off and on: with image mode on, known
// emoji clusters measure as image cells even when no image placement
// exists, so any text fallback must occupy the same cell count.
func TestMessagesPaneRowsExactWidth(t *testing.T) {
	for _, mode := range []struct {
		name   string
		active bool
	}{
		{"images-off", false},
		{"images-on", true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			emoji.SetImageMode(mode.active, 2)
			defer emoji.SetImageMode(false, 2)

			// A VS16 emoji in the text and in a reaction pill: as a
			// glyph it renders one cell in kitty while unstripped
			// measurement says two — the misaligned-border bug this
			// file pins.
			msgs := append(goldenMessages(), messages.MessageItem{
				TS:        goldenTS(2 * time.Minute),
				UserID:    "U2",
				UserName:  "bob",
				Text:      "careful ⚠️ with this",
				Timestamp: "9:02 AM",
				DateStr:   "2026-03-15",
				Reactions: []messages.ReactionItem{
					{Emoji: "warning", Count: 2, UserIDs: []string{"U1"}},
				},
			})

			a := newGoldenApp(t, withSize(120, 34))
			a.messagepane.SetMessages(msgs)
			_ = a.View()

			const innerW = 118 // msgWidth-2 at 120 cols
			view := a.messagepane.ViewBare(28, innerW)
			lines := strings.Split(view, "\n")
			for i, l := range lines {
				if w := ansi.StringWidth(l); w != innerW {
					t.Errorf("[%s] row %d width = %d, want %d: %q",
						mode.name, i, w, innerW, ansi.Strip(l))
				}
			}
		})
	}
}
