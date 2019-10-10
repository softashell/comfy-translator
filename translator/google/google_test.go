package google

import "testing"

func Test_cleanText(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"do n't",
			args{
				"Do n't bother me",
			},
			"Don't bother me",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanText(tt.args.text); got != tt.want {
				t.Errorf("cleanText() = %v, want %v", got, tt.want)
			}
		})
	}
}
