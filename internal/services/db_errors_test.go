package services

import (
	"errors"
	"testing"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "duplicate key", err: errors.New("duplicate key value violates unique constraint \"clients_email_key\""), want: true},
		{name: "unique constraint", err: errors.New("unique constraint failed"), want: true},
		{name: "other error", err: errors.New("insert failed"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolation(tt.err)
			if got != tt.want {
				t.Fatalf("isUniqueViolation(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

