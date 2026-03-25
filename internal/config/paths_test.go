package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name    string
		confdir string
		input   string
		want    string
	}{
		{
			name:    "home prefix",
			confdir: "/etc",
			input:   "~/foo",
			want:    filepath.Join(home, "foo"),
		},
		{
			name:    "relative prefix",
			confdir: "/etc/dirwatch",
			input:   "./bar",
			want:    filepath.Join("/etc/dirwatch", "bar"),
		},
		{
			name:    "absolute path",
			confdir: "/etc",
			input:   "/abs/path",
			want:    "/abs/path",
		},
		{
			name:    "empty string",
			confdir: "/etc",
			input:   "",
			want:    "",
		},
		{
			name:    "single char",
			confdir: "/etc",
			input:   "x",
			want:    "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.confdir, tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q, %q) = %q, want %q", tt.confdir, tt.input, got, tt.want)
			}
		})
	}
}
