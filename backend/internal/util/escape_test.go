package util_test

import (
	"testing"

	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
)

func TestEscapeMarkdown_basicEscapes(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"a_b", "a\\_b"},
		{"(test)", "\\(test\\)"},
		{"dash - dash", "dash \\- dash"},
	}
	for _, tt := range tests {
		if got := util.EscapeMarkdown(tt.in); got != tt.want {
			t.Errorf("EscapeMarkdown(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEscapeMarkdown_codeFence(t *testing.T) {
	input := "```\nfoo - bar\n```"
	want := "```\nfoo \\- bar\n```"
	if got := util.EscapeMarkdown(input); got != want {
		t.Errorf("EscapeMarkdown fence = %q, want %q", got, want)
	}
}
