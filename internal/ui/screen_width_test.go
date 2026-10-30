package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/gammons/slk/internal/emoji"
)

// TestScreenLinesExactWidth pins the terminal-level contract: every line
// View() emits must be exactly the terminal width. A short line lets the
// previous frame's cells (or the terminal default background) survive to
// the right of the new content — visibly, a border fragment one
// character out of place; a long one wraps and shifts every later row.
func TestScreenLinesExactWidth(t *testing.T) {
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

			a := newGoldenApp(t, withSize(120, 34))
			a.messagepane.SetMessages(goldenMessages())
			v := a.View()
			lines := strings.Split(v.Content, "\n")
			for i, l := range lines {
				if w := ansi.StringWidth(l); w != 120 {
					t.Errorf("[%s] screen line %d width = %d, want 120: %q",
						mode.name, i, w, ansi.Strip(l))
				}
			}
		})
	}
}
