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
			name: "with custom encoder",
			args: args{
				encoder: encoder,
			},
			want: AssertionTable{
				{Input: 0, Output: "aza", Comment: "Lower boundary"},
				{Input: 13, Output: "oko", Comment: "Mid-range (50%)"},
				{Input: 26, Output: "iti", Comment: "Upper boundary (single word)"},
			},
			wantErr: false,
		},
		{
			name: "nil encoder uses defaults",
			args: args{
				encoder: nil,
			},
			want: AssertionTable{
				{Input: 0, Output: "babab-babab", Comment: "Lower boundary"},
				{Input: 2147483647, Output: "luzuz-zuzuz", Comment: "Mid-range (50%)"},
				{Input: 4294967295, Output: "zuzuz-zuzuz", Comment: "Upper boundary (single word)"},
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

func TestGenerateSuggestionsWithCustom(t *testing.T) {
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
		encoder           *p.PhoneticEncoder
		customValues      []p.PositiveInt
		includeBoundaries bool
	}
	tests := []struct {
		name    string
		args    args
		want    AssertionTable
		wantErr bool
	}{
		{
			name: "custom values only",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{5, 10, 20},
				includeBoundaries: false,
			},
			want: AssertionTable{
				{Input: 5, Output: "aki", Comment: "Custom check for 5"},
				{Input: 10, Output: "ozo", Comment: "Custom check for 10"},
				{Input: 20, Output: "izi", Comment: "Custom check for 20"},
			},
			wantErr: false,
		},
		{
			name: "boundaries only",
			args: args{
				encoder:           encoder,
				customValues:      nil,
				includeBoundaries: true,
			},
			want: AssertionTable{
				{Input: 0, Output: "aza", Comment: "Lower boundary"},
				{Input: 13, Output: "oko", Comment: "Mid-range (50%)"},
				{Input: 26, Output: "iti", Comment: "Upper boundary (single word)"},
			},
			wantErr: false,
		},
		{
			name: "boundaries and custom values",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{7, 15},
				includeBoundaries: true,
			},
			want: AssertionTable{
				{Input: 0, Output: "aza", Comment: "Lower boundary"},
				{Input: 13, Output: "oko", Comment: "Mid-range (50%)"},
				{Input: 26, Output: "iti", Comment: "Upper boundary (single word)"},
				{Input: 7, Output: "ato", Comment: "Custom check for 7"},
				{Input: 15, Output: "ota", Comment: "Custom check for 15"},
			},
			wantErr: false,
		},
		{
			name: "custom value exceeds capacity",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{100},
				includeBoundaries: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "neither boundaries nor custom values",
			args: args{
				encoder:           encoder,
				customValues:      nil,
				includeBoundaries: false,
			},
			want:    AssertionTable{},
			wantErr: false,
		},
		{
			name: "localhost IP address (127.0.0.1 = 2130706433) with default config",
			args: args{
				encoder:           nil, // Use default ProQuint encoder
				customValues:      []p.PositiveInt{2130706433},
				includeBoundaries: false,
			},
			want: AssertionTable{
				{Input: 2130706433, Output: "lusab-babad", Comment: "Custom check for 2130706433"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSuggestionsWithCustom(tt.args.encoder, tt.args.customValues, tt.args.includeBoundaries)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSuggestionsWithCustom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSuggestionsWithCustom() = %v, want %v", got, tt.want)
			}
		})
	}
}
