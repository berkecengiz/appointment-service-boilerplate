package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int
		wantError bool
	}{
		{
			name:      "valid positive integer",
			input:     "42",
			want:      42,
			wantError: false,
		},
		{
			name:      "valid zero",
			input:     "0",
			want:      0,
			wantError: false,
		},
		{
			name:      "valid negative integer",
			input:     "-10",
			want:      -10,
			wantError: false,
		},
		{
			name:      "valid large integer",
			input:     "123456789",
			want:      123456789,
			wantError: false,
		},
		{
			name:      "invalid non-numeric",
			input:     "abc",
			want:      0,
			wantError: true,
		},
		{
			name:      "valid with trailing letters (parses first number)",
			input:     "123abc",
			want:      123,
			wantError: false,
		},
		{
			name:      "invalid empty string",
			input:     "",
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid decimal",
			input:     "12.34",
			want:      12,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.Equal(t, 0, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

