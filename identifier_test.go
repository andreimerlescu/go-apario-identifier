package go_apario_identifier

import (
	`reflect`
	`testing`
	`time`
)

func TestIdentifier_String(t *testing.T) {
	type fields struct {
		Year     int16
		Fragment Fragment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test one",
			fields: fields{
				Year:     2024,
				Fragment: CodeFragment("ABC123"),
			},
			want: "2024ABC123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Identifier{
				Year:     tt.fields.Year,
				Fragment: tt.fields.Fragment,
			}
			if got := i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_Path(t *testing.T) {
	type fields struct {
		Year     int16
		Fragment Fragment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test one",
			fields: fields{
				Year:     2024,
				Fragment: CodeFragment("ABC123"),
			},
			want: "2024/A/B/C1/23",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Identifier{
				Year:     tt.fields.Year,
				Fragment: tt.fields.Fragment,
			}
			if got := i.Path(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIdentifierURL(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *Identifier
	}{
		{
			name: "get Identifier{} from idoread.com/valet/documents/2024/ABC123DEF",
			args: args{
				path: "idoread.com/valet/documents/2024/ABC123DEF",
			},
			want: &Identifier{
				Instance:  []rune("idoread.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("ABC123DEF"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIdentifierURL(tt.args.path)
			if !reflect.DeepEqual(got.Instance, tt.want.Instance) {
				t.Errorf("ParseIdentifierURL().Instance = %v, want %v", got.Instance, tt.want.Instance)
			}
			if !reflect.DeepEqual(got.Table, tt.want.Table) {
				t.Errorf("ParseIdentifierURL().Table = %v, want %v", got.Table, tt.want.Table)
			}
			if !reflect.DeepEqual(got.Year, tt.want.Year) {
				t.Errorf("ParseIdentifierURL().Year = %v, want %v", got.Year, tt.want.Year)
			}
			if !reflect.DeepEqual(got.Fragment, tt.want.Fragment) {
				t.Errorf("ParseIdentifierURL().Fragment = %v, want %v", got.Fragment, tt.want.Fragment)
			}
		})
	}
}

func TestIdentifier_UUID(t *testing.T) {
	type fields struct {
		Instance  []rune
		Concierge []rune
		Table     []rune
		Year      int16
		Fragment  Fragment
		e         error
		eat       time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "UUID 1",
			fields: fields{
				Instance:  []rune("idoread.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("ABCDEFGHI"),
			},
			want: "105.100.111.114.101.97.100.46.99.111.109-118.97.108.101.116-100.111.99.117.109.101.110.116.115-2024-65.66.67.68.69.70.71.72.73",
		},
		{
			name: "UUID 2",
			fields: fields{
				Instance:  []rune("docs.projectminnesota.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("ALKHDA"),
			},
			want: "100.111.99.115.46.112.114.111.106.101.99.116.109.105.110.110.101.115.111.116.97.46.99.111.109-118.97.108.101.116-100.111.99.117.109.101.110.116.115-2024-65.76.75.72.68.65",
		},
		{
			name: "UUID 3",
			fields: fields{
				Instance:  []rune("projectapario.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("SD9DKLH93"),
			},
			want: "112.114.111.106.101.99.116.97.112.97.114.105.111.46.99.111.109-118.97.108.101.116-100.111.99.117.109.101.110.116.115-2024-83.68.57.68.75.76.72.57.51",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Identifier{
				Instance:  tt.fields.Instance,
				Concierge: tt.fields.Concierge,
				Table:     tt.fields.Table,
				Year:      tt.fields.Year,
				Fragment:  tt.fields.Fragment,
				e:         tt.fields.e,
				eat:       tt.fields.eat,
			}
			if got := i.UUID(); got != tt.want {
				t.Errorf("UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_IsRemote(t *testing.T) {
	type fields struct {
		Instance  []rune
		Concierge []rune
		Table     []rune
		Year      int16
		Fragment  Fragment
		e         error
		eat       time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "local identifier",
			fields: fields{
				Instance:  []rune{},
				Concierge: []rune{},
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("ABCDEF"),
			},
			want: false,
		},
		{
			name: "remote identifier",
			fields: fields{
				Instance:  []rune("idoread.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("ABCDEF"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Identifier{
				Instance:  tt.fields.Instance,
				Concierge: tt.fields.Concierge,
				Table:     tt.fields.Table,
				Year:      tt.fields.Year,
				Fragment:  tt.fields.Fragment,
				e:         tt.fields.e,
				eat:       tt.fields.eat,
			}
			if got := i.IsRemote(); got != tt.want {
				t.Errorf("IsRemote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_Online(t *testing.T) {
	type fields struct {
		Instance  []rune
		Concierge []rune
		Table     []rune
		Year      int16
		Fragment  Fragment
		Version   *Version
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test one",
			fields: fields{
				Instance:  []rune("idoread.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  IntegerFragment(1),
				Version:   &Version{0, 0, 1},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Identifier{
				Instance:  tt.fields.Instance,
				Concierge: tt.fields.Concierge,
				Table:     tt.fields.Table,
				Year:      tt.fields.Year,
				Fragment:  tt.fields.Fragment,
				Version:   tt.fields.Version,
			}
			if got := i.Online(); got != tt.want {
				t.Errorf("Online() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifierForUUID(t *testing.T) {
	type args struct {
		uuid string
	}
	tests := []struct {
		name string
		args args
		want *Identifier
	}{
		{
			name: "test one",
			args: args{
				uuid: "112.114.111.106.101.99.116.97.112.97.114.105.111.46.99.111.109-118.97.108.101.116-100.111.99.117.109.101.110.116.115-2024-83.68.57.68.75.76.72.57.51",
			},
			want: &Identifier{
				Instance:  []rune("projectapario.com"),
				Concierge: []rune("valet"),
				Table:     []rune("documents"),
				Year:      2024,
				Fragment:  []rune("SD9DKLH93"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IdentifierForUUID(tt.args.uuid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IdentifierForUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
