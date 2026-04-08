//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	phonid "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

var (
	encoder       *phonid.PhoneticEncoder
	currentConfig *phonid.PhonidConfig
)

const jsSafeIntegerMax = 9007199254740991

// loadConfig accepts TOML content as string and initializes the encoder
func loadConfig(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return map[string]interface{}{
			"error": "expected 1 argument: config content (TOML string)",
		}
	}

	configContent := args[0].String()

	// Parse phonidrc TOML format with preflight checks
	phonidCfg, preflightChecks, err := phonid.ParsePhonidRCLenient(configContent)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to parse TOML: %v", err),
		}
	}

	// Create encoder (skip preflight for initial load)
	enc, err := phonid.NewPhoneticEncoderSkipPreflight(phonidCfg)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to create encoder: %v", err),
		}
	}

	// Store globally
	encoder = enc
	currentConfig = phonidCfg

	// Validate preflight checks if provided
	var preflightError string
	if len(preflightChecks) > 0 {
		if err := encoder.ValidatePreflight(preflightChecks); err != nil {
			preflightError = fmt.Sprintf("Preflight validation failed: %v", err)
		}
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Config loaded successfully",
	}

	if preflightError != "" {
		result["preflightError"] = preflightError
	}

	return result
}

// encode converts a number to a phonetic string
func encode(this js.Value, args []js.Value) interface{} {
	if encoder == nil {
		return map[string]interface{}{
			"error": "encoder not initialized - load config first",
		}
	}

	if len(args) != 1 {
		return map[string]interface{}{
			"error": "expected 1 argument: number to encode",
		}
	}

	// Handle both int and string inputs (string for BigInt support)
	var num phonid.PositiveInt
	arg := args[0]

	switch arg.Type() {
	case js.TypeNumber:
		f := arg.Float()
		if f < 0 || f != float64(int64(f)) {
			return map[string]interface{}{
				"error": "number must be a non-negative integer",
			}
		}
		if f > jsSafeIntegerMax {
			return map[string]interface{}{
				"error": "number exceeds JavaScript safe integer range; pass as decimal or 0x hex string",
			}
		}
		num = phonid.NewPositiveInt(int64(f))
	case js.TypeString:
		// String representation for BigInt
		var err error
		num, err = phonid.ParsePositiveInt(arg.String())
		if err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("invalid number: %v", err),
			}
		}
	default:
		return map[string]interface{}{
			"error": "argument must be a number or string",
		}
	}

	result, err := encoder.Encode(num)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("encoding failed: %v", err),
		}
	}

	return map[string]interface{}{
		"result": result,
	}
}

// decode converts a phonetic string back to a number
func decode(this js.Value, args []js.Value) interface{} {
	if encoder == nil {
		return map[string]interface{}{
			"error": "encoder not initialized - load config first",
		}
	}

	if len(args) != 1 {
		return map[string]interface{}{
			"error": "expected 1 argument: string to decode",
		}
	}

	str := args[0].String()
	result, err := encoder.Decode(str)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("decoding failed: %v", err),
		}
	}

	// Try to return as int64 if it fits, otherwise as string
	if num, ok := result.Int64(); ok {
		return map[string]interface{}{
			"result":      num,
			"resultStr":   result.String(),
			"fitsInInt64": true,
		}
	}

	return map[string]interface{}{
		"result":      result.String(),
		"resultStr":   result.String(),
		"fitsInInt64": false,
	}
}

// generateSuggestions creates preflight check suggestions
func generateSuggestions(this js.Value, args []js.Value) interface{} {
	if encoder == nil {
		return map[string]interface{}{
			"error": "encoder not initialized - load config first",
		}
	}

	if currentConfig == nil {
		return map[string]interface{}{
			"error": "config not available",
		}
	}

	// Generate full TOML config with valid preflight checks
	configStr, err := preflight.SuggestConfig(encoder, currentConfig)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to generate config: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"config":  configStr,
	}
}

// getVersion returns the phonid version
func getVersion(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"version": "1.0.0",
		"wasm":    true,
	}
}

// loadDefaultConfig loads the ProQuint-compatible default config
func loadDefaultConfig(this js.Value, args []js.Value) interface{} {
	// Use default ProQuint config
	enc, err := phonid.NewPhoneticEncoderSkipPreflight(&phonid.ProQuintConfig)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to create encoder: %v", err),
		}
	}

	encoder = enc

	return map[string]interface{}{
		"success": true,
		"message": "Default ProQuint-compatible config loaded",
	}
}

// getRandom generates a random number/encoding pair
func getRandom(this js.Value, args []js.Value) interface{} {
	if encoder == nil {
		return map[string]interface{}{
			"error": "encoder not initialized - load config first",
		}
	}

	assertion, err := preflight.GetRandom(encoder)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to generate random: %v", err),
		}
	}

	// Try to return as int64 if it fits, otherwise as string
	if num, ok := assertion.Input.Value.Int64(); ok {
		return map[string]interface{}{
			"number":      num,
			"numberStr":   assertion.Input.Value.String(),
			"encoding":    assertion.Output,
			"fitsInInt64": true,
		}
	}

	return map[string]interface{}{
		"number":      assertion.Input.Value.String(),
		"numberStr":   assertion.Input.Value.String(),
		"encoding":    assertion.Output,
		"fitsInInt64": false,
	}
}

func main() {
	c := make(chan struct{})

	// Register functions to be called from JavaScript
	js.Global().Set("phonidLoadConfig", js.FuncOf(loadConfig))
	js.Global().Set("phonidEncode", js.FuncOf(encode))
	js.Global().Set("phonidDecode", js.FuncOf(decode))
	js.Global().Set("phonidGenerateSuggestions", js.FuncOf(generateSuggestions))
	js.Global().Set("phonidGetVersion", js.FuncOf(getVersion))
	js.Global().Set("phonidGetRandom", js.FuncOf(getRandom))
	js.Global().Set("phonidLoadDefaultConfig", js.FuncOf(loadDefaultConfig))

	// Signal that WASM is ready
	js.Global().Call("dispatchEvent", js.Global().Get("CustomEvent").New("phonidReady"))

	fmt.Println("Phonid WASM initialized - functions available:")
	fmt.Println("  - phonidLoadConfig(tomlString)")
	fmt.Println("  - phonidLoadDefaultConfig()")
	fmt.Println("  - phonidEncode(numberOrString)")
	fmt.Println("  - phonidDecode(string)")
	fmt.Println("  - phonidGenerateSuggestions()")
	fmt.Println("  - phonidGetRandom()")
	fmt.Println("  - phonidGetVersion()")

	<-c // Keep the program running
}
