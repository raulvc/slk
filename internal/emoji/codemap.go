package emoji

import "strings"

// CodeMap returns the standard-emoji shortcode→glyph map, keyed by
// ":name:" (all Slack aliases) with bare-glyph values (no trailing
// space). It is slk's replacement for kyokomi/emoji's CodeMap(), sourced
// from the bundled iamcal/emoji-data table (slacknames_gen.go).
//
// The returned map is shared and must not be mutated by callers (all
// current callers only read it).
func CodeMap() map[string]string { return slackCodeMap }

// Sprint substitutes Slack-style :shortcode: sequences in s with their
// Unicode glyph, appending a trailing space after each — byte-for-byte
// compatible with the kyokomi/emoji Sprint() it replaces. Unknown
// shortcodes and non-shortcode text (including ASCII emoticons like ":)")
// pass through unchanged.
func Sprint(s string) string {
	return shortcodeRe.ReplaceAllStringFunc(s, func(match string) string {
		if glyph, ok := slackCodeMap[match]; ok {
			return StripVariationSelectors(glyph) + " "
		}
		return match
	})
}

// StripVariationSelectors removes U+FE0F (emoji-presentation selector)
// from s. slk renders emoji as terminal GLYPHS in a monospace font, and
// the presentation selector is the classic width trap: terminals with a
// color-emoji fallback draw the pair two cells wide, while kitty with a
// mono font (e.g. JetBrainsMono Nerd Font Mono) draws the narrow text
// glyph — one cell. Go width tables (ansi, uniseg) count the pair as 2,
// so rows containing the pair pad for 2 but render 1 and the pane
// border prints one character to the left. Rendering the bare base
// codepoint makes measure and render agree at 1 cell on every terminal.
// Kitty-image placements are unaffected: they carry the emoji's CDN URL
// (which keeps its canonical fe0f form), not the glyph.
func StripVariationSelectors(s string) string {
	return strings.ReplaceAll(s, "\uFE0F", "")
}
