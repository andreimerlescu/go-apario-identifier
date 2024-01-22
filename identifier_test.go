package go_apario_identifier

import (
	`testing`
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
