package go_apario_identifier

import (
	`os`
	`testing`
	`time`
)

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
			temp, tmpErr := os.MkdirTemp("", tt.args.databasePrefixPath)
			if tmpErr != nil {
				t.Errorf("os.MkdirTemp() error = %v, wantErr %v", tmpErr, tt.wantErr)
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

			code := string(got.Fragment)
			if len(code) != tt.args.length {
				t.Errorf("code length invalid ; expected %d got %d", tt.args.length, len(code))
				return
			}
		})
	}
}
