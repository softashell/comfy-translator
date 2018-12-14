package main

import "testing"

func Test_matchWhitespace(t *testing.T) {
	type args struct {
		text   string
		source string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Leading whitespace",
			args: args{
				text: "a",
				source: "  	b",
			},
			want: "  	a",
		},
		{
			name: "Trailing whitespace",
			args: args{
				text: "a",
				source: "b	  ",
			},
			want: "a	  ",
		},
		{
			name: "Leading and trailing whitespace",
			args: args{
				text: "a",
				source: "  	b	  ",
			},
			want: "  	a	  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchWhitespace(tt.args.text, tt.args.source); got != tt.want {
				t.Errorf("matchWhitespace() = %q, want %q", got, tt.want)
			}
		})
	}
}
