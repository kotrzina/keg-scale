package hook

import "testing"

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
