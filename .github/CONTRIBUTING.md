# Contributing to Phonid

Thank you for your interest in contributing to Phonid! 🎉

## Contributing Public Presets

Public presets are shared configurations that other users can easily load and use. They're displayed in the WASM demo and available via GitHub raw URLs.

### Requirements

1. **Generated Configuration**: Must be created using `phonid preflight --suggest`
2. **Naming Convention**: Must match `.*.phonidrc.toml` (e.g., `.mypreset.phonidrc.toml`)
3. **Location**: Must be in `public_presets/` directory
4. **Signed Commits**: All commits must be GPG or SSH signed
5. **Preflight Checks**: All assertions must pass
6. **Distinct Use Case**: Should serve a purpose different from existing presets

### Step-by-Step Guide

#### 1. Use the Helper Script (Recommended)

The easiest way to create a preset:

```bash
./scripts/create-preset.sh
```

This will guide you through the entire process!
