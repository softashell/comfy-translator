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
						input:  "ニコニコローン 即日融資‼",
						output: "NicoNico loan the same day loan! ",
					},
					{
						input:  "繁華街雑居ビル地下3階",
						output: "Downtown Town Building Underground 3 Floor",
					},
				},
			},
			want: []responsePair{
				{
					input:  "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
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
						input:  "ニコニコローン 即日融資‼",
						output: "NicoNico loan the same day loan! ",
					},
					{
						input:  "繁華街雑居ビル地下3階",
						output: "Downtown Town Building Underground 3 Floor",
					},
					{
						input:  "A",
						output: "A",
					},
				},
			},
			want: []responsePair{
				{
					input:  "ニコニコローン 即日融資‼繁華街雑居ビル地下3階",
					output: "NicoNico loan the same day loan! Downtown Town Building Underground 3 Floor",
				},
				{
					input:  "A",
					output: "A",
				},
			},
		},
		{
			name: "input split on server side, keep first and last two items",
			args: args{
				input: []inputObject{
					{
						req: &translator.Request{
							Text: "A",
						},
					},
					{
						req: &translator.Request{
							Text: "もう人気Ｎｏ．１といっても過言じゃないんじゃないかな",
						},
					},
					{
						req: &translator.Request{
							Text: "あっあっふあぁぁっ",
						},
					},
					{
						req: &translator.Request{
							Text: "あっあっふあぁぁっ2",
						},
					},
				},
				output: []responsePair{
					{
						input:  "A",
						output: "A",
					},
					{
						input:  "もう人気Ｎｏ．",
						output: "It is already popular No.",
					},
					{
						input:  "１といっても過言じゃないんじゃないかな",
						output: "It might not be an exaggeration to say 1",
					},
					{
						input:  "あっあっふあぁぁっ",
						output: "Aaaaaaaaaa",
					},
					{
						input:  "あっあっふあぁぁっ2",
						output: "Aaaaaaaaaa2",
					},
				},
			},
			want: []responsePair{
				{
					input:  "A",
					output: "A",
				},								
				{
					input:  "もう人気Ｎｏ．１といっても過言じゃないんじゃないかな",
					output: "It is already popular No.It might not be an exaggeration to say 1",
				},
				{
					input:  "あっあっふあぁぁっ",
					output: "Aaaaaaaaaa",
				},
				{
					input:  "あっあっふあぁぁっ2",
					output: "Aaaaaaaaaa2",
				},
			},
			
		},
		{
			name: "input split on server side, keep first two items",
			args: args{
				input: []inputObject{
					{
						req: &translator.Request{
							Text: "あっあっふあぁぁっ",
						},
					},
					{
						req: &translator.Request{
							Text: "あっあっふあぁぁっ2",
						},
					},
					{
						req: &translator.Request{
							Text: "もう人気Ｎｏ．１といっても過言じゃないんじゃないかな",
						},
					},
				},
				output: []responsePair{
					{
						input:  "あっあっふあぁぁっ",
						output: "Aaaaaaaaaa",
					},
					{
						input:  "あっあっふあぁぁっ2",
						output: "Aaaaaaaaaa2",
					},
					{
						input:  "もう人気Ｎｏ．",
						output: "It is already popular No.",
					},
					{
						input:  "１といっても過言じゃないんじゃないかな",
						output: "It might not be an exaggeration to say 1",
					},

				},
			},
			want: []responsePair{
				{
					input:  "あっあっふあぁぁっ",
					output: "Aaaaaaaaaa",
				},
				{
					input:  "あっあっふあぁぁっ2",
					output: "Aaaaaaaaaa2",
				},						
				{
					input:  "もう人気Ｎｏ．１といっても過言じゃないんじゃないかな",
					output: "It is already popular No.It might not be an exaggeration to say 1",
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
