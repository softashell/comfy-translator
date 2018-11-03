package google

import (
	"reflect"
	"testing"

	"gitgud.io/softashell/comfy-translator/translator"
)

func Test_mergeOutput(t *testing.T) {
	type args struct {
		input  []inputObject
		output []responsePair
	}
	tests := []struct {
		name string
		args args
		want []responsePair
	}{
		{
			name: "input split on server side, delete last item",
			args: args{
				input: []inputObject{
					{
						req: &translator.Request{
							Text: "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
						},
					},
				},
				output: []responsePair{
					{
						input: "ニコニコローン 即日融資‼",
						output: "NicoNico loan the same day loan! ",
					},
					{
						input: "繁華街雑居ビル地下3階",
						output: "Downtown Town Building Underground 3 Floor",
					},
				},
			},
			want: []responsePair{
				{
					input: "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
					output: "NicoNico loan the same day loan! Downtown Town Building Underground 3 Floor",
				},
			},
		},
		{
			name: "input split on server side, keep last item",
			args: args{
				input: []inputObject{
					{
						req: &translator.Request{
							Text: "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
						},
					},
					{
						req: &translator.Request{
							Text: "A",
						},
					},
				},
				output: []responsePair{
					{
						input: "ニコニコローン 即日融資‼",
						output: "NicoNico loan the same day loan! ",
					},
					{
						input: "繁華街雑居ビル地下3階",
						output: "Downtown Town Building Underground 3 Floor",
					},
					{
						input: "A",
						output: "A",
					},
				},
			},
			want: []responsePair{
				{
					input: "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
					output: "NicoNico loan the same day loan! Downtown Town Building Underground 3 Floor",
				},
				{
					input: "A",
					output: "A",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeOutput(tt.args.input, tt.args.output); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
