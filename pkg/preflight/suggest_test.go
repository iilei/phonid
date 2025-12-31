package preflight_test

import (
	"reflect"
	"testing"

	p "github.com/iilei/phonid/pkg"
	. "github.com/iilei/phonid/pkg/preflight"
)

func TestGenerateSuggestions(t *testing.T) {
	placeholderMap := p.PlaceholderMap{p.Vowel: p.RuneSet{'a', 'o', 'i'}, p.Consonant: p.RuneSet{'z', 'k', 't'}}

	config := &p.PhonidConfig{
		Patterns:     []string{"VCV"},
		Placeholders: placeholderMap,
	}
	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Errorf("NewPhoneticEncoderLenient() error: %v", err)
	}
	type args struct {
		encoder *p.PhoneticEncoder
	}
	tests := []struct {
		name    string
		args    args
		want    AssertionTable
		wantErr bool
	}{
		{
			name: "asdf",
			args: args{
				encoder: encoder,
			},
			want: AssertionTable{
				{Input: 0, Output: "aza", Comment: "Lower boundary"},
				{Input: 26, Output: "iti", Comment: "Upper boundary (single word)"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSuggestions(tt.args.encoder)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSuggestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSuggestions() = %v, want %v", got, tt.want)
			}
		})
	}
}
