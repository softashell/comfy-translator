package google

import (
	"reflect"
	"testing"
)

func Test_decodeResponse(t *testing.T) {
	type args struct {
		resp string
	}
	tests := []struct {
		name    string
		args    args
		want    []responsePair
		wantErr bool
	}{
		{
			name: "basic test",
			args: args{
				resp: `[[["With me this thing","これで俺の事",null,null,3]],null,"ja"]`,
			},
			want: []responsePair{
				{
					"これで俺の事",
					"With me this thing",
				},
			},
			wantErr: false,
		},
		{
			name: "basic test 2",
			args: args{
				resp: `[[["output","input",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]]],null,"ja"]`,
			},
			want: []responsePair{
				{
					"input",
					"output",
				},
			},
			wantErr: false,
		},
		{
			name: "multiline test",
			args: args{
				resp: `[[["テストテスト1.5\n","test test1.5\n",null,null,3],["テスト2\n","test2\n",null,null,3],["テスト3\n","test3\n",null,null,3],["テスト4テスト4.5\n","test4 test4.5\n",null,null,3],["テスト5テスト","test5 testing",null,null,3],[null,null,"Tesuto tesuto 1. 5 Tesuto 2 tesuto 3 tesuto 4 tesuto 4. 5 Tesuto 5 tesuto"]],null,"en",null,null,[["test test1.5",null,[["テストテスト1.5",0,true,false],["テストtest1.5",0,true,false]],[[0,12]],"test test1.5",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test2",null,[["テスト2",0,true,false],["TEST2",0,true,false]],[[0,5]],"test2",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test3",null,[["テスト3",0,true,false],["TEST3",0,true,false]],[[0,5]],"test3",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test4 test4.5",null,[["テスト4テスト4.5",0,true,false],["TEST4のtest4.5",0,true,false]],[[0,13]],"test4 test4.5",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test5 testing",null,[["テスト5テスト",0,true,false],["TEST5テスト",0,true,false]],[[0,13]],"test5 testing",0,0]],1,["test \u003cb\u003e\u003ci\u003etest 1\u003c/i\u003e\u003c/b\u003e.5 test2 test3 test4 test4.5 test5 testing","test test 1.5\ntest2\ntest3\ntest4 test4.5\ntest5 testing",[1],null,null,false],[["en"],null,[1],["en"]]]`,
			},
			want: []responsePair{
				{
					"test test1.5",
					"テストテスト1.5",
				},
				{
					"test2",
					"テスト2",
				},
				{
					"test3",
					"テスト3",
				},
				{
					"test4 test4.5",
					"テスト4テスト4.5",
				},
				{
					"test5 testing",
					"テスト5テスト",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeResponse(tt.args.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanResponseText(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "newline with trailing space",
			args: args{
				s: "ま、嬢ちゃんの身体は一つだから仕方ないな\n ",
			},
			want: "ま、嬢ちゃんの身体は一つだから仕方ないな",
		},
		{
			name: "keep leading space",
			args: args{
				s: " こんなスケスケ衣装で寝るぐらいだから、 ",
			},
			want: " こんなスケスケ衣装で寝るぐらいだから、",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanResponseText(tt.args.s); got != tt.want {
				t.Errorf("cleanResponseText() = %v, want %v", got, tt.want)
			}
		})
	}
}
