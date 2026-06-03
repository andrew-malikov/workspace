package containers

import "testing"

func TestHasRunning(t *testing.T) {
	tests := []struct {
		name string
		out  []byte
		want bool
	}{
		{
			name: "empty output",
			want: false,
		},
		{
			name: "empty json array",
			out:  []byte("[]\n"),
			want: false,
		},
		{
			name: "json object",
			out:  []byte(`{"Name":"api"}`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRunning(tt.out)
			if got != tt.want {
				t.Fatalf("unexpected result: got %t want %t", got, tt.want)
			}
		})
	}
}
