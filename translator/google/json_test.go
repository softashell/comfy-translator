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
		{
			name:  "new google format",
			input: `[[["Start taking off all clothes","服をすべて脱がす起動",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]]],null,"ja"]`,
			want:  `[["Start taking off all clothes","服をすべて脱がす起動"]]`,
		},
		{
			name:  "new google format long",
			input: `"[[["It's a very well-dressed shop assistant\\n","おやおや随分と大胆な格好の店員さんだねぇ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["No. 1 brother and son\\n","喧嘩No1兄貴と子分\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Town 01\\n","町01\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["shop\\n","お店\\n",null,null,1],["Church interior 3\\n","教会内部3\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Library 2\\n","図書館2\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Guru\\n","教祖\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Inside the church\\n","教会内部\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Leonardo's Road\\n","レオナルドの道\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Tess\\n","てすとわざ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It was a town famous for training mentals\\n","それはメンタルを鍛えることで有名な町だった\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Church interior 2\\n","教会内部2\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Visit the people in the West Building\\n","西館にいる者を訪ねよ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It is not very spacious, but it is a comfortable space\\n","あまり広くは無いですが、快適な空間ですよ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I do not have enough alcohol!\\n","おらおら!酒が足らないぞッ!もっともってこい!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Girls who are being bullied at school want to change themselves\\n","学校でいじめられている女の子が自分を変えるべく\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Normal ending town\\n","ノーマルエンディング用町\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It is a man of Bukabuka\\n","ブカブカの男物だ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It is Tetsuto\\n","てすとですね\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["That's enough!\\n","もう十分!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The trial version is up to here !!! Look forward to the release of the product version !!!\\n","体験版はココまでです!!!製品版の発売をお楽しみに!!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Water play\\n","水遊び\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Are you sure the event will proceed?\\n","イベントが進みますがよろしいですか?\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I bump into my skirt and it turns over\\n","ぶつかってスカートがめくれる\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["There was no such thing\\n","そんなことは無かったね\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Do you mean the swimsuit?\\n","水着の意味を成しているのか?と疑いたくなる極小サイズ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The story of a girl without such\\n","そんなさえない女の子の物語\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Heavenly God, our God, let me admire\\n","天にまします我らの神、我にあがめさせたまえ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The trial version is up to here !! Have fun with the product version !!!\\n","体験版はココまでです!!製品版をお楽しみに!!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I'm a tailor\\n","私は仕立て屋の\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Here\\n","ここかぁ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Speaking of decent things here, it's about drinks\\n","ココでまともなものと言えば飲み物ぐらいだ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It's too horny to come to the store and talk to me\\n","店にまで来て僕に話しかけるなんて余程淫乱なんだね\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Be careful if you create a retrospective room because you are not managing it at a common event\\n","コモンイベントで管理していないから回想部屋を作るなら注意\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Thank you!!\\n","よろしくお願いします!!\\n",null,null,1],["Heart of the Buddha\\n","御仏の心を\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Wind Oh oh oh !!!\\n","ウィンドおああああ!!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Old fun\\n","旧楽しいです\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["General\\n","大将\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["See you then\\n","じゃあ、確認しますぜ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Oh, it's a cute girl again\\n","あら?またかわいい子ちゃんね\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["ointment\\n","軟膏\\n",null,null,2],["Souh Souh\\n","スーハースーハー\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Thank you!\\n","ありがとうございます!\\n",null,null,1],["The bookshelf here isn't so strong, absolutely beyond\\n","ここの本棚はそんなに丈夫じゃないし、絶対に向こうにも\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The resident removed all Mika's clothes when continuing so\\n","そう続けると住人はミカの服をすべて脱がした\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Have you heard about the contents of the training that you will do this time?\\n","今回やって貰う修行の内容って聞いてる?\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["White flag surrender!\\n","白旗降参!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Even so, it's totally crazy\\n","それにしても全く、けしからん格好ですな\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I heard that waitress is busy here\\n","ここの店はウェイトレスがブスって聞いてたけど\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Mika\\n","ミカ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Certainly this formula may not be wrong\\n","確かにこの計算式間違ってないかもな\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Changing the clothes because the pants were stolen\\n","パンツが盗まれたから着替えなおすイベント\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It is unpleasant to hear\\n","聞いてて不愉快だ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Uoh oh oh !!\\n","うぉおおおお!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It is Mika that came for training!\\n","修行のためにやってきましたミカです!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It's really fun\\n","真楽しいです\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Another voice !!\\n","もう一声!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I was tired by the time I came here, so I took a break\\n","今日はここに来るまでに疲れてしまったので少し休んでも\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["It was correct\\n","おしかったな\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Ordinary uniforms everywhere\\n","どこにでもある普通の制服\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Window's !!!\\n","ウインドウッ!!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Chant\\n","唱えなさい\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["An event in which pants are stolen by grandfather\\n","じいさんにパンツを盗まれるイベント\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Librarian\\n","図書館員\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Please cooperate with the library that everyone can enjoy happily\\n","皆が楽しく利用できる図書館にご協力お願いします\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The manager's food is bad\\n","店長の料理は下手くそさ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The hot spring here is very effective and popular\\n","ここの温泉は効能がすごくて人気なんだよ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["※Please※\\n","※お願い※\\n",null,null,1],["One page of such a day\\n","そんなある日の1ページ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I will do my best\\n","絶対に手をどかしてやる\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Is it so strange to come in a swimsuit?\\n","水着で来るのがそんなに変なことですか?\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I am a gurus here\\n","私はここで教祖をやっておる者じゃ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["4 if 2\\n","4なら2\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["No entry allowed for overall cleaning!\\n","全面清掃のため立ち入り禁止!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I am sorry to talk suddenly,\\n","突然話かけて申し訳ない、\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Mika did not think in front of her thinking\\n","ミカは考えごとをして目の前が見えていなかった\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Oh, it was a treat\\n","あっ、ご馳走様でした\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Hey\\n","ちょっとッ\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Voice is small! One more time!\\n","声が小さい!もう一回だッ!!\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Six months have passed since Mika finished training\\n","ミカが修行を終えて半年の月日が流れた\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I want to\\n","たく\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["Or\\n","オル\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["I can not concentrate on my work anymore\\n","もうダメだ、仕事に集中できない\\n",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]],["The top ball so far, a little idiot for sale","ここまでの上玉、売り物にするにはちょっとばかし",null,null,3,null,null,null,[[["7b78faa1329566192a016d05fd7850f4","ja_en_2018q4.md"]]]]],null,"ja"]"`,
			want:  `[[["It's a very well-dressed shop assistant\\n","おやおや随分と大胆な格好の店員さんだねぇ\\n"],["No. 1 brother and son\\n","喧嘩No1兄貴と子分\\n"],["Town 01\\n","町01\\n"],["shop\\n","お店\\n"],["Church interior 3\\n","教会内部3\\n"],["Library 2\\n","図書館2\\n"],["Guru\\n","教祖\\n"],["Inside the church\\n","教会内部\\n"],["Leonardo's Road\\n","レオナルドの道\\n"],["Tess\\n","てすとわざ\\n"],["It was a town famous for training mentals\\n","それはメンタルを鍛えることで有名な町だった\\n"],["Church interior 2\\n","教会内部2\\n"],["Visit the people in the West Building\\n","西館にいる者を訪ねよ\\n"],["It is not very spacious, but it is a comfortable space\\n","あまり広くは無いですが、快適な空間ですよ\\n"],["I do not have enough alcohol!\\n","おらおら!酒が足らないぞッ!もっともってこい!!\\n"],["Girls who are being bullied at school want to change themselves\\n","学校でいじめられている女の子が自分を変えるべく\\n"],["Normal ending town\\n","ノーマルエンディング用町\\n"],["It is a man of Bukabuka\\n","ブカブカの男物だ\\n"],["It is Tetsuto\\n","てすとですね\\n"],["That's enough!\\n","もう十分!\\n"],["The trial version is up to here !!! Look forward to the release of the product version !!!\\n","体験版はココまでです!!!製品版の発売をお楽しみに!!!\\n"],["Water play\\n","水遊び\\n"],["Are you sure the event will proceed?\\n","イベントが進みますがよろしいですか?\\n"],["I bump into my skirt and it turns over\\n","ぶつかってスカートがめくれる\\n"],["There was no such thing\\n","そんなことは無かったね\\n"],["Do you mean the swimsuit?\\n","水着の意味を成しているのか?と疑いたくなる極小サイズ\\n"],["The story of a girl without such\\n","そんなさえない女の子の物語\\n"],["Heavenly God, our God, let me admire\\n","天にまします我らの神、我にあがめさせたまえ\\n"],["The trial version is up to here !! Have fun with the product version !!!\\n","体験版はココまでです!!製品版をお楽しみに!!!\\n"],["I'm a tailor\\n","私は仕立て屋の\\n"],["Here\\n","ここかぁ\\n"],["Speaking of decent things here, it's about drinks\\n","ココでまともなものと言えば飲み物ぐらいだ\\n"],["It's too horny to come to the store and talk to me\\n","店にまで来て僕に話しかけるなんて余程淫乱なんだね\\n"],["Be careful if you create a retrospective room because you are not managing it at a common event\\n","コモンイベントで管理していないから回想部屋を作るなら注意\\n"],["Thank you!!\\n","よろしくお願いします!!\\n"],["Heart of the Buddha\\n","御仏の心を\\n"],["Wind Oh oh oh !!!\\n","ウィンドおああああ!!!\\n"],["Old fun\\n","旧楽しいです\\n"],["General\\n","大将\\n"],["See you then\\n","じゃあ、確認しますぜ\\n"],["Oh, it's a cute girl again\\n","あら?またかわいい子ちゃんね\\n"],["ointment\\n","軟膏\\n"],["Souh Souh\\n","スーハースーハー\\n"],["Thank you!\\n","ありがとうございます!\\n"],["The bookshelf here isn't so strong, absolutely beyond\\n","ここの本棚はそんなに丈夫じゃないし、絶対に向こうにも\\n"],["The resident removed all Mika's clothes when continuing so\\n","そう続けると住人はミカの服をすべて脱がした\\n"],["Have you heard about the contents of the training that you will do this time?\\n","今回やって貰う修行の内容って聞いてる?\\n"],["White flag surrender!\\n","白旗降参!\\n"],["Even so, it's totally crazy\\n","それにしても全く、けしからん格好ですな\\n"],["I heard that waitress is busy here\\n","ここの店はウェイトレスがブスって聞いてたけど\\n"],["Mika\\n","ミカ\\n"],["Certainly this formula may not be wrong\\n","確かにこの計算式間違ってないかもな\\n"],["Changing the clothes because the pants were stolen\\n","パンツが盗まれたから着替えなおすイベント\\n"],["It is unpleasant to hear\\n","聞いてて不愉快だ\\n"],["Uoh oh oh !!\\n","うぉおおおお!!\\n"],["It is Mika that came for training!\\n","修行のためにやってきましたミカです!\\n"],["It's really fun\\n","真楽しいです\\n"],["Another voice !!\\n","もう一声!!\\n"],["I was tired by the time I came here, so I took a break\\n","今日はここに来るまでに疲れてしまったので少し休んでも\\n"],["It was correct\\n","おしかったな\\n"],["Ordinary uniforms everywhere\\n","どこにでもある普通の制服\\n"],["Window's !!!\\n","ウインドウッ!!!\\n"],["Chant\\n","唱えなさい\\n"],["An event in which pants are stolen by grandfather\\n","じいさんにパンツを盗まれるイベント\\n"],["Librarian\\n","図書館員\\n"],["Please cooperate with the library that everyone can enjoy happily\\n","皆が楽しく利用できる図書館にご協力お願いします\\n"],["The manager's food is bad\\n","店長の料理は下手くそさ\\n"],["The hot spring here is very effective and popular\\n","ここの温泉は効能がすごくて人気なんだよ\\n"],["※Please※\\n","※お願い※\\n"],["One page of such a day\\n","そんなある日の1ページ\\n"],["I will do my best\\n","絶対に手をどかしてやる\\n"],["Is it so strange to come in a swimsuit?\\n","水着で来るのがそんなに変なことですか?\\n"],["I am a gurus here\\n","私はここで教祖をやっておる者じゃ\\n"],["4 if 2\\n","4なら2\\n"],["No entry allowed for overall cleaning!\\n","全面清掃のため立ち入り禁止!\\n"],["I am sorry to talk suddenly,\\n","突然話かけて申し訳ない、\\n"],["Mika did not think in front of her thinking\\n","ミカは考えごとをして目の前が見えていなかった\\n"],["Oh, it was a treat\\n","あっ、ご馳走様でした\\n"],["Hey\\n","ちょっとッ\\n"],["Voice is small! One more time!\\n","声が小さい!もう一回だッ!!\\n"],["Six months have passed since Mika finished training\\n","ミカが修行を終えて半年の月日が流れた\\n"],["I want to\\n","たく\\n"],["Or\\n","オル\\n"],["I can not concentrate on my work anymore\\n","もうダメだ、仕事に集中できない\\n"],["The top ball so far, a little idiot for sale","ここまでの上玉、売り物にするにはちょっとばかし"]]`,
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
