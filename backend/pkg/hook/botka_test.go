package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhatsApp_SanitizeCommand(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{
			command: "/cenik",
			want:    "cenik",
		},
		{
			command: "/ cenik ",
			want:    "cenik",
		},
		{
			command: "/cenik 1",
			want:    "cenik 1",
		},
		{
			command: "/CENIK",
			want:    "cenik",
		},
		{
			command: "/BEČKA",
			want:    "becka",
		},
		{
			command: "!cep Plzen",
			want:    "cep plzen",
		},
	}

	b := Botka{}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := b.sanitizeCommand(tt.command); got != tt.want {
				t.Errorf("sanitizeCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAmountFromQrPaymentCommand(t *testing.T) {
	tests := []struct {
		command string
		err     bool
		want    int
	}{
		{
			command: "/qr 100",
			err:     false,
			want:    100,
		},
		{
			command: "/qr 275",
			err:     false,
			want:    275,
		},
		{
			command: "qr 50",
			err:     false,
			want:    50,
		},
		{
			command: "QR 50",
			err:     false,
			want:    50,
		},
		{
			command: "Qr 50",
			err:     false,
			want:    50,
		},
		{
			command: "/qr",
			err:     true,
			want:    0,
		},
		{
			command: "/QR",
			err:     true,
			want:    0,
		},
		{
			command: "qr",
			err:     true,
			want:    0,
		},
		{
			command: "qr and more text",
			err:     true,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, err := parseAmountFromQrPaymentCommand(tt.command)
			if (err != nil) != tt.err {
				t.Errorf("parseAmountFromQrPaymentCommand() error = %v, wantErr %v", err, tt.err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
