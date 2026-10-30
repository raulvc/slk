package messages

import (
	"github.com/gammons/slk/internal/bubbles/entrylist"
	"github.com/gammons/slk/internal/ui/styles"
)

// themeListStyles maps the current theme onto the shared list substrate's
// styles. Both content panes call it at the top of every View, so a
// theme change reaches the window padding, scrollbar and selection
// highlight on the next frame, exactly when the panes' own styles.* reads
// do.
func themeListStyles() entrylist.Styles {
	return entrylist.Styles{
		Background: styles.Background,
		Track:      styles.Border,
		// Thumb in the muted color, not Primary: a bright full-block
		// column one cell inside the border reads as a misprinted
		// border rather than a scrollbar (the thumb is one column
		// wide by design — scrollbar.Overlay draws it at width-1).
		Thumb:     styles.TextMuted,
		Selection: styles.SelectionStyle(),
	}
}

// ThemeListStyles is themeListStyles for the thread pane.
func ThemeListStyles() entrylist.Styles { return themeListStyles() }
