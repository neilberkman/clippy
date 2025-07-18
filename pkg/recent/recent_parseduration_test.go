package recent

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:  "empty string returns default 5 minutes",
			input: "",
			want:  5 * time.Minute,
		},
		{
			name:  "positive number as minutes",
			input: "10",
			want:  10 * time.Minute,
		},
		{
			name:    "negative number should error",
			input:   "-5",
			wantErr: true,
		},
		{
			name:  "standard duration format",
			input: "2h30m",
			want:  2*time.Hour + 30*time.Minute,
		},
		{
			name:    "negative duration should error",
			input:   "-2h",
			wantErr: true,
		},
		{
			name:  "seconds format",
			input: "30s",
			want:  30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}