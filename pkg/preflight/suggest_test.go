package preflight_test

import (
	"reflect"
	"testing"

	p "github.com/iilei/phonid/pkg"
	. "github.com/iilei/phonid/pkg/preflight"
)

const patternCVCCV = "CVCCV"

// makeTwoPatternBoundaries creates boundary assertions for two patterns.
func makeTwoPatternBoundaries(pattern2 string) AssertionTable {
	output121 := "okkoo"
	output242 := "ittii"
	if pattern2 == patternCVCCV {
		output121 = "kokko"
		output242 = "titti"
	}

	return AssertionTable{
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(0)}, Output: "aza", Comment: "Lower boundary (VCV)"},
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(13)}, Output: "oko", Comment: "Mid-range (VCV)"},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(26)},
			Output:  "iti",
			Comment: "Upper boundary (VCV)",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
			Output:  "aza",
			Comment: "Lower boundary (" + pattern2 + ")",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(121)},
			Output:  output121,
			Comment: "Mid-range (" + pattern2 + ")",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(242)},
			Output:  output242,
			Comment: "Upper boundary (" + pattern2 + ")",
		},
	}
}

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

	multiPatternConfig := &p.PhonidConfig{
		Patterns:     []string{"VCV", "CVCCV"},
		Placeholders: placeholderMap,
	}
	multiPatternEncoder, err := p.NewPhoneticEncoderLenient(multiPatternConfig)
	if err != nil {
		t.Errorf("NewPhoneticEncoderLenient() error for multiPatternConfig: %v", err)
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
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(0)}, Output: "aza", Comment: "Lower boundary (VCV)"},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(13)}, Output: "oko", Comment: "Mid-range (VCV)"},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(26)},
					Output:  "iti",
					Comment: "Upper boundary (VCV)",
				},
			},
			wantErr: false,
		},
		{
			name: "with multiple patterns",
			args: args{
				encoder: multiPatternEncoder,
			},
			want:    makeTwoPatternBoundaries(patternCVCCV),
			wantErr: false,
		},
		{
			name: "nil encoder uses defaults",
			args: args{
				encoder: nil,
			},
			want: AssertionTable{
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
					Output:  "babab-babab",
					Comment: "Lower boundary (CVCVCXCVCVC)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(2147483647)},
					Output:  "luzuz-zuzuz",
					Comment: "Mid-range (CVCVCXCVCVC)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(4294967295)},
					Output:  "zuzuz-zuzuz",
					Comment: "Upper boundary (CVCVCXCVCVC)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(2130706433)},
					Output:  "lusab-babad",
					Comment: "localhost IP address (127.0.0.1 = 2130706433)",
				},
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
		Patterns:     []string{"VCV", "VCCVV"},
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
				customValues:      []p.PositiveInt{p.NewPositiveInt(5), p.NewPositiveInt(10), p.NewPositiveInt(20)},
				includeBoundaries: false,
			},
			want: AssertionTable{
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(5)}, Output: "aki", Comment: "Custom check for 5"},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(10)}, Output: "ozo", Comment: "Custom check for 10"},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(20)}, Output: "izi", Comment: "Custom check for 20"},
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
			want:    makeTwoPatternBoundaries("VCCVV"),
			wantErr: false,
		},
		{
			name: "boundaries and custom values",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{p.NewPositiveInt(7), p.NewPositiveInt(15)},
				includeBoundaries: true,
			},
			want: AssertionTable{
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(0)}, Output: "aza", Comment: "Lower boundary (VCV)"},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(13)}, Output: "oko", Comment: "Mid-range (VCV)"},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(26)},
					Output:  "iti",
					Comment: "Upper boundary (VCV)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
					Output:  "aza",
					Comment: "Lower boundary (VCCVV)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(121)},
					Output:  "okkoo",
					Comment: "Mid-range (VCCVV)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(242)},
					Output:  "ittii",
					Comment: "Upper boundary (VCCVV)",
				},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(7)}, Output: "ato", Comment: "Custom check for 7"},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(15)}, Output: "ota", Comment: "Custom check for 15"},
			},
			wantErr: false,
		},
		{
			name: "custom value exceeds capacity",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{p.NewPositiveInt(300)},
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
