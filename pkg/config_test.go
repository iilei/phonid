package phonid

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := &Config{
		Phonetic: &PhonidConfig{}, // defaults are applied when validatePattern is called
		Shuffle:  &ShuffleConfig{32, 0, 0},
	}

	tests := []struct {
		name    string
		want    Config
		wantErr bool
	}{
		{"sensible defaults", *cfg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig(): wantErr=%v, got error=%v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(got.Phonetic, tt.want.Phonetic) {
				t.Errorf(
					"NewConfigWithOptions().Phonetic = %v, want %v",
					got.Phonetic,
					tt.want.Phonetic,
				)
			}
			if !reflect.DeepEqual(got.Shuffle, tt.want.Shuffle) {
				t.Errorf(
					"NewConfigWithOptions().Shuffle = %v, want %v",
					got.Shuffle,
					tt.want.Shuffle,
				)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Phonetic *PhonidConfig
		Shuffle  *ShuffleConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Phonetic: tt.fields.Phonetic,
				Shuffle:  tt.fields.Shuffle,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfigWithOptions(t *testing.T) {
	type args struct {
		opts []ConfigOption
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigWithOptions(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigWithOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithBitWidth(t *testing.T) {
	type args struct {
		bitWidth int
	}
	tests := []struct {
		name string
		args args
		want ConfigOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithBitWidth(tt.args.bitWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithBitWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithRounds(t *testing.T) {
	type args struct {
		rounds int
	}
	tests := []struct {
		name string
		args args
		want ConfigOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithRounds(tt.args.rounds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithRounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithSeed(t *testing.T) {
	type args struct {
		seed uint64
	}
	tests := []struct {
		name string
		args args
		want ConfigOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithSeed(tt.args.seed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithShuffle(t *testing.T) {
	type args struct {
		shuffle *ShuffleConfig
	}
	tests := []struct {
		name string
		args args
		want ConfigOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithShuffle(tt.args.shuffle); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithShuffle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithPhonetic(t *testing.T) {
	type args struct {
		phonetic *PhonidConfig
	}
	tests := []struct {
		name string
		args args
		want ConfigOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPhonetic(tt.args.phonetic); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithPhonetic() = %v, want %v", got, tt.want)
			}
		})
	}
}
