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

func Test_cleanJson(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "gtx single response",
			input: `[[["With me this thing","これで俺の事",null,null,3]],null,"ja"]`,
			want:  `[["With me this thing","これで俺の事"]]`,
		},
		{
			name:  "t multiline response",
			input: `[[["テストテスト1.5\n","test test1.5\n",null,null,3],["テスト2\n","test2\n",null,null,3],["テスト3\n","test3\n",null,null,3],["テスト4テスト4.5\n","test4 test4.5\n",null,null,3],["テスト5テスト","test5 testing",null,null,3],[null,null,"Tesuto tesuto 1. 5 Tesuto 2 tesuto 3 tesuto 4 tesuto 4. 5 Tesuto 5 tesuto"]],null,"en",null,null,[["test test1.5",null,[["テストテスト1.5",0,true,false],["テストtest1.5",0,true,false]],[[0,12]],"test test1.5",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test2",null,[["テスト2",0,true,false],["TEST2",0,true,false]],[[0,5]],"test2",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test3",null,[["テスト3",0,true,false],["TEST3",0,true,false]],[[0,5]],"test3",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test4 test4.5",null,[["テスト4テスト4.5",0,true,false],["TEST4のtest4.5",0,true,false]],[[0,13]],"test4 test4.5",0,0],["\n",null,null,[[0,1]],"\n",0,0],["test5 testing",null,[["テスト5テスト",0,true,false],["TEST5テスト",0,true,false]],[[0,13]],"test5 testing",0,0]],1,["test \u003cb\u003e\u003ci\u003etest 1\u003c/i\u003e\u003c/b\u003e.5 test2 test3 test4 test4.5 test5 testing","test test 1.5\ntest2\ntest3\ntest4 test4.5\ntest5 testing",[1],null,null,false],[["en"],null,[1],["en"]]]`,
			want:  `[["テストテスト1.5\n","test test1.5\n"],["テスト2\n","test2\n"],["テスト3\n","test3\n"],["テスト4テスト4.5\n","test4 test4.5\n"],["テスト5テスト","test5 testing"]]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanJson(tt.input); got != tt.want {
				t.Errorf("cleanJson() = %v, want %v", got, tt.want)
			}
		})
	}
}
