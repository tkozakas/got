package telegram

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
			name:    "standardSeconds",
			input:   "30s",
			want:    30 * time.Second,
			wantErr: false,
		},
		{
			name:    "standardMinutes",
			input:   "5m",
			want:    5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "standardHours",
			input:   "2h",
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "customDayFormatOneDay",
			input:   "1d",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "customDayFormatTwoDays",
			input:   "2d",
			want:    48 * time.Hour,
			wantErr: false,
		},
		{
			name:    "combinedHoursMinutes",
			input:   "1h30m",
			want:    1*time.Hour + 30*time.Minute,
			wantErr: false,
		},
		{
			name:    "combinedDaysHours",
			input:   "1d2h",
			want:    26 * time.Hour,
			wantErr: false,
		},
		{
			name:    "combinedDaysHoursMinutes",
			input:   "1d2h30m",
			want:    24*time.Hour + 2*time.Hour + 30*time.Minute,
			wantErr: false,
		},
		{
			name:    "combinedDaysHoursMinutesSeconds",
			input:   "1d2h30m45s",
			want:    24*time.Hour + 2*time.Hour + 30*time.Minute + 45*time.Second,
			wantErr: false,
		},
		{
			name:    "invalidFormat",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "emptyString",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalidAlphabetic",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
