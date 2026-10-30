package sidebar

import (
	"strings"
	"testing"
)

// fakeSectionsProvider serves a fixed Slack-section list (fakeProvider
// in cursor_reorder_test.go covers the reorder case; this one carries
// emoji for the label tests).
type fakeSectionsProvider struct{ metas []SectionMeta }

func (f fakeSectionsProvider) Ready() bool                         { return true }
func (f fakeSectionsProvider) OrderedSlackSections() []SectionMeta { return f.metas }

// TestSectionLabelOverride pins the section_labels config: an override
// by section ID or by current name renames the header; shortcodes in
// the override resolve to glyphs; other sections are untouched.
func TestSectionLabelOverride(t *testing.T) {
	m := New([]ChannelItem{
		{ID: "C1", Name: "general", Type: "channel", Section: "L0SEC1"},
		{ID: "C2", Name: "random", Type: "channel", Section: "L0SEC2"},
	})
	m.SetSectionsProvider(fakeSectionsProvider{metas: []SectionMeta{
		{ID: "L0SEC1", Name: "papo sobre grafana", Emoji: "raphael-estiloso", Type: "standard"},
		{ID: "L0SEC2", Name: "Company", Type: "standard"},
	}})

	// Override by ID, carrying a resolvable unicode shortcode.
	m.SetSectionLabels(map[string]string{
		"L0SEC1": ":calendar: grafana talk",
	})
	view := m.View(20, 30)

	if !strings.Contains(view, "grafana talk") || strings.Contains(view, ":calendar:") {
		t.Errorf("ID-keyed override did not render; view:\n%s", view)
	}
	if strings.Contains(view, "raphael-estiloso") {
		t.Errorf("custom emoji shortcode leaked into the header:\n%s", view)
	}

	// Override by current NAME for the other section.
	m.SetSectionLabels(map[string]string{"Company": "Work stuff"})
	view = m.View(20, 30)
	if !strings.Contains(view, "Work stuff") {
		t.Errorf("name-keyed override did not render; view:\n%s", view)
	}
}

// No labels configured: everything renders as Slack sent it (with the
// emoji prefix attempt for custom emoji staying literal).
func TestSectionLabelOverrideAbsent(t *testing.T) {
	m := New([]ChannelItem{
		{ID: "C1", Name: "general", Type: "channel", Section: "L0SEC1"},
	})
	m.SetSectionsProvider(fakeSectionsProvider{metas: []SectionMeta{
		{ID: "L0SEC1", Name: "papo sobre grafana", Emoji: "raphael-estiloso", Type: "standard"},
	}})
	view := m.View(20, 30)
	if !strings.Contains(view, "papo sobre grafana") {
		t.Errorf("section name missing with no override configured; view:\n%s", view)
	}
}
