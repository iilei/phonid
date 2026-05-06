package preflight_test

import (
	"reflect"
	"testing"

	p "github.com/iilei/phonid/pkg"
	. "github.com/iilei/phonid/pkg/preflight"
)

const (
	patternCVCVCVC   = "CVCVCVC"
	patternVCVCV     = "VCVCV"
	placeholderOther = "other"

	suggestOutputKokokok = "kokokok"
	suggestOutputTititit = "tititit"
	suggestOutputItiti   = "ititi"
	suggestOutputAzaza   = "azaza"
	suggestOutputOkoko   = "okoko"

	suggestLowerVCVCV = "Lower boundary (VCVCV)"
	suggestMidVCVCV   = "Mid-range (VCVCV)"
	suggestUpperVCVCV = "Upper boundary (VCVCV)"
)

// makeTwoPatternBoundaries creates boundary assertions for two patterns.
func makeTwoPatternBoundaries(pattern2 string) AssertionTable {
	output1093 := suggestOutputKokokok // CVCVCVC starts at 243
	output2186 := suggestOutputTititit
	if pattern2 != patternCVCVCVC {
		output1093 = placeholderOther
		output2186 = placeholderOther
	}

	return AssertionTable{
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(0)}, Output: suggestOutputAzaza, Comment: suggestLowerVCVCV},
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(121)}, Output: suggestOutputOkoko, Comment: suggestMidVCVCV},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(242)},
			Output:  suggestOutputItiti,
			Comment: suggestUpperVCVCV,
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
			Output:  suggestOutputAzaza,
			Comment: "Lower boundary (" + pattern2 + ")",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(1093)},
			Output:  output1093,
			Comment: "Mid-range (" + pattern2 + ")",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(2186)},
			Output:  output2186,
			Comment: "Upper boundary (" + pattern2 + ")",
		},
	}
}

func TestGenerateSuggestions(t *testing.T) {
	placeholderMap := p.PlaceholderMap{p.Vowel: p.RuneSet{'a', 'o', 'i'}, p.Consonant: p.RuneSet{'z', 'k', 't'}}

	config := &p.PhonidConfig{
		Patterns:     []string{patternVCVCV},
		Placeholders: placeholderMap,
	}
	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Errorf("NewPhoneticEncoderLenient() error: %v", err)
	}

	multiPatternConfig := &p.PhonidConfig{
		Patterns:     []string{patternVCVCV, patternCVCVCVC},
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
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
					Output:  suggestOutputAzaza,
					Comment: suggestLowerVCVCV,
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(121)},
					Output:  suggestOutputOkoko,
					Comment: suggestMidVCVCV,
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(242)},
					Output:  suggestOutputItiti,
					Comment: suggestUpperVCVCV,
				},
			},
			wantErr: false,
		},
		{
			name: "with multiple patterns",
			args: args{
				encoder: multiPatternEncoder,
			},
			want:    makeTwoPatternBoundaries(patternCVCVCVC),
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
		Patterns:     []string{patternVCVCV, patternCVCVCVC},
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
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(5)}, Output: "azaki", Comment: "Custom check for 5"},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(10)},
					Output:  "azozo",
					Comment: "Custom check for 10",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(20)},
					Output:  "azizi",
					Comment: "Custom check for 20",
				},
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
			want:    makeTwoPatternBoundaries("CVCVCVC"),
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
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
					Output:  suggestOutputAzaza,
					Comment: suggestLowerVCVCV,
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(121)},
					Output:  suggestOutputOkoko,
					Comment: suggestMidVCVCV,
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(242)},
					Output:  suggestOutputItiti,
					Comment: suggestUpperVCVCV,
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
					Output:  "azaza",
					Comment: "Lower boundary (CVCVCVC)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(1093)},
					Output:  suggestOutputKokokok,
					Comment: "Mid-range (CVCVCVC)",
				},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(2186)},
					Output:  suggestOutputTititit,
					Comment: "Upper boundary (CVCVCVC)",
				},
				{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(7)}, Output: "azato", Comment: "Custom check for 7"},
				{
					Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(15)},
					Output:  "azota",
					Comment: "Custom check for 15",
				},
			},
			wantErr: false,
		},
		{
			name: "custom value exceeds capacity",
			args: args{
				encoder:           encoder,
				customValues:      []p.PositiveInt{p.NewPositiveInt(3000)}, // Exceeds CVCVCVC capacity (2186)
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
