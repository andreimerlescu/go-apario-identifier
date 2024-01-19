package go_apario_identifier

import (
	`errors`
	`fmt`
	`os`
	`reflect`
	`strings`
	`testing`
	`time`
)

func TestIdentifierPath(t *testing.T) {
	type args struct {
		identifier string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "id 2023ABCDEFG",
			args: args{
				identifier: "2023ABCDEFG",
			},
			want: "2023/A/B/CD/EFG",
		},
		{
			name: "id 3033YSUDHSCHDSKHSIEHF",
			args: args{
				identifier: "3033YSUDHSCHDSKHSIEHF",
			},
			want: "3033/Y/S/UD/HSC/HDSKH/SIEHF",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IdentifierPath(tt.args.identifier); got != tt.want {
				t.Errorf("IdentifierPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIdentifier(t *testing.T) {
	type args struct {
		identifier string
	}
	tests := []struct {
		name    string
		args    args
		want    *Identifier
		wantErr bool
	}{
		{
			name: "test invalid year identifier",
			args: args{
				identifier: "3333ABC123DEF",
			},
			want: &Identifier{
				Year: int16(3333),
				Code: "abc123def",
			},
			wantErr: true,
		},
		{
			name: "test valid year identifier",
			args: args{
				identifier: "2024ABC123DEF",
			},
			want: &Identifier{
				Year: int16(2024),
				Code: "abc123def",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIdentifier(tt.args.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseIdentifier() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newToken(t *testing.T) {
	type args struct {
		length   int
		attempts int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "generate new token 6 chars 17 attempts",
			args: args{
				length:   6,
				attempts: 17,
			},
			wantLen: 10,
			wantErr: false,
		},
		{
			name: "generate new token 6 chars 0 attempts expected err",
			args: args{
				length:   6,
				attempts: 0,
			},
			wantLen: 10,
			wantErr: true,
		},
		{
			name: "generate new token 6 chars -1 attempts expected err",
			args: args{
				length:   6,
				attempts: -1,
			},
			wantLen: 10,
			wantErr: true,
		},
		{
			name: "generate new token 0 chars -1 attempts expected err",
			args: args{
				length:   0,
				attempts: -1,
			},
			wantLen: 4,
			wantErr: true,
		},
		{
			name: "generate new token -1 chars -1 attempts expected err",
			args: args{
				length:   -1,
				attempts: -1,
			},
			wantLen: 4,
			wantErr: true,
		},
		{
			name: "generate new token 29 chars 1 attempt",
			args: args{
				length:   29,
				attempts: 1,
			},
			wantLen: 33,
			wantErr: false,
		},
		{
			name: "generate new token 30 chars -1 attempt expected err",
			args: args{
				length:   30,
				attempts: -1,
			},
			wantLen: 34,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newToken(tt.args.length, tt.args.attempts)
			if (tt.wantErr && err == nil) || (!tt.wantErr && err != nil) {
				t.Errorf("newToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !tt.wantErr && err == nil {
				if len(got.String()) != tt.wantLen {
					t.Errorf("newToken() got = %v, want %v", len(got.String()), tt.wantLen)
				}
			}

		})
	}
}

func Test_generateIdentifier(t *testing.T) {
	type args struct {
		databasePrefixPath string
		length             int
		attempts           int
	}
	tests := []struct {
		name    string
		args    args
		want    *Identifier
		wantErr bool
	}{
		{
			name: "generate and verify new identifier on filesystem",
			args: args{
				databasePrefixPath: "users.db",
				length:             9,
				attempts:           12,
			},
			want:    &Identifier{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp, tmp_err := os.MkdirTemp("", tt.args.databasePrefixPath)
			if tmp_err != nil {
				t.Errorf("os.MkdirTemp() error = %v, wantErr %v", tmp_err, tt.wantErr)
				return
			}

			got, err := generateIdentifier(temp, tt.args.length, tt.args.attempts)

			err = os.RemoveAll(temp)
			if err != nil {
				t.Errorf("os.RemoveAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("generateIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Year != int16(time.Now().UTC().Year()) {
				t.Errorf("years mismatched wantErr %v", tt.wantErr)
				return
			}

			code := strings.ReplaceAll(got.String(), string(rune(time.Now().UTC().Year())), ``)
			if len(code) != tt.args.length {
				t.Errorf("code length invalid ; expected %d got %d", tt.args.length, len(code))
				return
			}
		})
	}
}

func TestNewIdentifier(t *testing.T) {
	tmpDir, dirErr := os.MkdirTemp("", "tmp-users.db")
	if dirErr != nil {
		t.Errorf("os.MkdirTemp() returned unexpected err %v", dirErr)
		return
	}
	for i := 3; i <= 69; i++ {
		identifier, newErr := NewIdentifier(tmpDir, i*1, i*2, i*3)
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("%d. os.RemoveAll() returned unexpected err %v", i, err)
			return
		}
		if newErr != nil {
			t.Errorf("%d. NewIdentifier() returned unexpected err %v", i, newErr)
			return
		}
		if len(identifier.Code) != i {
			t.Errorf("%d. identifier.Code returned unexpected err %v", i, errors.New(fmt.Sprintf("identifier.Code length %d != %d", len(identifier.Code), i)))
			return
		}
	}

}
