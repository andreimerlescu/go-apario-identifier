package go_apario_identifier

import (
	`errors`
	`fmt`
	`os`
	`reflect`
	`testing`
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
			name: "test 3333 invalid identifier",
			args: args{
				identifier: "3333ABC123DEF",
			},
			want: &Identifier{
				Year:     int16(3333),
				Fragment: CodeFragment("ABC123DEF"),
			},
			wantErr: false,
		},
		{
			name: "test 2024 valid identifier",
			args: args{
				identifier: "2024ABC123DEF",
			},
			want: &Identifier{
				Year:     int16(2024),
				Fragment: CodeFragment("ABC123DEF"),
			},
			wantErr: false,
		},
		{
			name: "test long 2024 valid identifier",
			args: args{
				identifier: "2024ABC123DEFHIJKLMNOPQ",
			},
			want: &Identifier{
				Year:     int16(2024),
				Fragment: CodeFragment("ABC123DEFHIJKLMNOPQ"),
			},
			wantErr: false,
		},
		{
			name: "test very long 2024 identifier",
			args: args{
				identifier: "2024ABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQ",
			},
			want: &Identifier{
				Year:     int16(2024),
				Fragment: CodeFragment("ABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQ"),
			},
			wantErr: false,
		},
		{
			name: "test very long invalid year identifier",
			args: args{
				identifier: "2024ABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQ",
			},
			want: &Identifier{
				Year:     int16(4444),
				Fragment: CodeFragment("ABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQABC123DEFHIJKLMNOPQ"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIdentifier(tt.args.identifier)
			if tt.wantErr {
				// want an err
				if err != nil {
					// got an err
					return
				}
				t.Errorf("ParseIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				// dont want an err
				if err != nil {
					// got an error
					t.Errorf("ParseIdentifier() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				// didnt get an error
				if !reflect.DeepEqual(got.Fragment, tt.want.Fragment) {
					t.Errorf("ParseIdentifier() got = %v, want %v", got, tt.want)
				}
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
		if i*2 > 29 { // 29 is max length of identifier ; +4 for year = 33 char identifier == (4^10)+(29^36) == 44 sexdecillion == 44 with 51 zeros after it aka 44,000,000,000,000,000,000,000,000,000,000,000,000,000(t),000(b),000(m),000(t),000(h)
			break
		}
		identifier, newErr := NewIdentifier(tmpDir, i*1, i*2, i*3)
		if newErr != nil {
			t.Errorf("%d. NewIdentifier() returned unexpected err %v", i, newErr)
			return
		}
		if len(identifier.Fragment) != i {
			t.Errorf("%d. identifier.Fragment returned unexpected err %v", i, errors.New(fmt.Sprintf("identifier.Fragment length %d != %d", len(identifier.Fragment), i)))
			return
		}
	}
	err := os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("os.RemoveAll() returned unexpected err %v", err)
		return
	}
}
