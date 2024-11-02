package scale

import (
	"testing"
)

func TestScale_GetPushResponse(t *testing.T) {
	cases := []struct {
		beersLeft int
		want      string
	}{
		{beersLeft: 0, want: "   0"},
		{beersLeft: 1, want: "   1"},
		{beersLeft: 10, want: "  10"},
		{beersLeft: 66, want: "  66"},
		{beersLeft: 100, want: " 100"},
		{beersLeft: 222, want: " 222"},
		{beersLeft: 1000, want: "1000"},
		{beersLeft: 9999, want: "9999"},
	}

	for _, tt := range cases {
		s := &Scale{
			beersLeft: tt.beersLeft,
		}

		got := s.GetPushResponse()
		if got != tt.want {
			t.Errorf("Scale.GetPushResponse() = %v, want %v", got, tt.want)
		}
	}
}
