// Package util provides helper utilities for the bot.
package util

import (
	"regexp"
	"strings"
)

// EscapeMarkdown prepares strings for safe MarkdownV2 output.
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)

	escaped := replacer.Replace(text)

	// Restore code fences and basic formatting markers
	escaped = strings.ReplaceAll(escaped, "\\`\\`\\`", "```")
	reInline := regexp.MustCompile(`\\` + "`" + `([^` + "`" + `]*?)\\` + "`")
	escaped = reInline.ReplaceAllString(escaped, "`$1`")
	reBold := regexp.MustCompile(`\\\*([^*]+?)\\\*`)
	escaped = reBold.ReplaceAllString(escaped, "*$1*")
	reItalic := regexp.MustCompile(`\\_([^_]+?)\\_`)
	escaped = reItalic.ReplaceAllString(escaped, "_$1_")

	return escaped
}
