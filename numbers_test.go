package go_apario_identifier

import (
	`fmt`
	`log`
	`os`
	`path/filepath`
	`reflect`
	`testing`
)

func TestIntegerFragment(t *testing.T) {
	for i := 9999; i <= 19999; i++ {
		log.Printf("i = %d ; fragment = %s", i, IntegerFragment(i))
	}
}

func TestCodeFragment(t *testing.T) {
	code := "ABC123"
	fragment := CodeFragment(code)
	if !reflect.DeepEqual(code, string(fragment)) {
		t.Errorf("failed to verify CodeFragment")
	}
}

func TestManifesting369(t *testing.T) {
	db, mkdirErr := os.MkdirTemp("", "gematria.db")
	if mkdirErr != nil {
		t.Errorf("os.MkdirTemp() returned err %v", mkdirErr)
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("failed to os.RemoveAll(%v) due to err %v", path, err)
		}
	}(db)

	valet := NewValet(db)
	cache, cacheErr := valet.GetCache(db)
	if cacheErr != nil {
		t.Errorf("valet.GetCache(db) returned err %v", cacheErr)
		return
	}

	codeEnglish := 1602
	codeJewish := 1028
	codeSimple := 267
	english := IntegerFragment(codeEnglish)
	jewish := IntegerFragment(codeJewish)
	simple := IntegerFragment(codeSimple)

	englishId, englishErr := english.ToIdentifier()
	if englishErr != nil {
		t.Errorf("received err = %v", englishErr)
		return
	}

	id, dir, idErr := cache.EnsureIdentifierDirectory(englishId.String())
	if idErr != nil {
		t.Errorf("cache.IdentifierCheck(%v) returned an err %v", englishId.String(), idErr)
		return
	}

	if !reflect.DeepEqual(id.Fragment, english) {
		t.Errorf("expected id.Fragment for english to be identical")
		return
	}

	info, infoErr := os.Stat(dir)
	if infoErr != nil {
		t.Errorf("os.Stat(%v) returned err %v", dir, infoErr)
		return
	}

	if !info.IsDir() {
		t.Errorf("!info.IsDir() indicated path %v is not a directory for identifier %v", dir, id.String())
	}

	_, setErr := valet.SetCache(db, englishId.String())
	if setErr != nil {
		t.Errorf("valet.SetCache(db, %v) return err %v", englishId.String(), setErr)
		return
	}

	if !valet.PathExists(filepath.Join(db, englishId.Path())) {
		t.Errorf("expecting directory %v to exist", englishId.Path())
		return
	}

	jewishId, jewishErr := jewish.ToIdentifier()
	if jewishErr != nil {
		t.Errorf("received err = %v", jewishErr)
		return
	}
	simpleId, simpleErr := simple.ToIdentifier()
	if simpleErr != nil {
		t.Errorf("received err = %v", simpleErr)
		return
	}

	log.Printf("storing %d = %s => %v\n", codeEnglish, englishId.String(), englishId.Path())
	log.Printf("storing %d = %s => %v\n", codeJewish, jewishId.String(), jewishId.Path())
	log.Printf("storing %d = %s => %v\n", codeSimple, simpleId.String(), simpleId.Path())
}

func TestNewFragment(t *testing.T) {
	type args struct {
		num int
	}
	tests := []struct {
		name string
		args args
		want Fragment
		code string
	}{
		{
			name: "test 3301",
			args: args{
				num: 3301,
			},
			want: Fragment{rune(48), rune(48), rune(48), rune(50), rune(74), rune(80)},
			code: "0002JP",
		},
		{
			name: "test 1602",
			args: args{
				num: 1602,
			},
			want: Fragment{rune(48), rune(48), rune(48), rune(49), rune(56), rune(73)},
			code: "00018I",
		},
		{
			name: "test 948384",
			args: args{
				num: 948384,
			},
			want: Fragment{rune(48), rune(48), rune(75), rune(66), rune(83), rune(48)},
			code: "00KBS0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntegerFragment(tt.args.num)
			if !reflect.DeepEqual(got, tt.want) || !reflect.DeepEqual(got.String(), tt.code) {
				t.Errorf("tt.args.num = %d ; got = %s ; want = %s", tt.args.num, got.String(), tt.want.String())
			} else {
				fmt.Printf("[SUCCESS] gematria value %d return identifier %v\n", tt.args.num, got.String())
			}
		})
	}
}
