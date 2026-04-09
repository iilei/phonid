# Phonid

> **Try it live:** Experiment with different configurations and encoding patterns in the [interactive playground](https://iilei.github.io/phonid/)

Phonid is a Go library for encoding and decoding numeric identifiers into **pronounceable, human-friendly fantasy words** while preserving **strict mathematical reversibility**.

This project is inspired by **Proquints**, especially the original paper, *Proquints: Identifiers that are Readable, Spellable, and Pronounceable*:
<https://arxiv.org/html/0901.4016>
<!-- archived: https://web.archive.org/web/20260401000408/https://arxiv.org/html/0901.4016 -->

Phonid generalizes that idea into a configurable system for different "phonetic languages" (e.g. Minion-like, Elvish-like) **without sacrificing bidirectional decodability**.

## What Phonid Guarantees

Phonid is intentionally strict:

* Every encoded identifier **must be uniquely decodable** (given the configuration and seed).
* No configuration option may introduce lossy or ambiguous transformations.
* All decoding decisions must be **deterministic and non-heuristic**.
* Performance should be predictable and close to constant time.

The output is playful, but the model is conservative and mathematically reversible.

## How Encoding Works

Phonid represents numbers as words of consonants (`C`) and vowels (`V`) using explicit templates.

### Explicit Templates

Each config defines a finite set of templates, for example:

* Length 3: `CVC`, `CVV`
* Length 5: `CVCCV`, `CVCVC`

A template determines:

* Which alphabet (consonant/vowel) is used at each position
* How many symbols each position can represent
* How values are packed and unpacked

Templates are never inferred; only declared templates are valid.

### Disjointness Rules

To keep decoding unambiguous, templates must be disjoint:

* No template may be a prefix of another template
* No shorter template may appear as a contiguous substring of a longer template
* Templates are validated at configuration load time

So decoding is simple and deterministic:

1. Determine the word length
2. Derive the C/V signature of the word
3. Perform an exact lookup of `(length, template)`

No backtracking, greedy matching, or heuristics are ever required.

## Configuration

Configurations allow:

* Defining alphabets (consonants, vowels)
* Declaring valid word templates per word length
* Selecting enabled word lengths

Configurations do not allow:

* Context-sensitive rules
* Any transformation that cannot be mathematically reversed


## Versioning and Stability

Phonid follows Semantic Versioning (SemVer).

> While the major version is `0.x.y`, **breaking changes may occur at any time**.
> Stability guarantees apply only after `v1.0.0`.

## Local Dev

Install GoReleaser:

```sh
go install github.com/goreleaser/goreleaser/v2@latest
```

Then follow the commands in **Quick Start (5 Commands)** below.

## Quick Start (5 Commands)

1. Build:

```sh
goreleaser build --snapshot --clean --single-target
```

2. Generate and save config:

```sh
./dist/phonid*/phonid preflight --suggest >| .phonidrc.tmp && mv .phonidrc.tmp .phonidrc.toml
```

3. Encode:

```sh
./dist/phonid*/phonid 4711
```


5. Encode a SHA-256 value (exact 256-bit preset):

```sh
./dist/phonid*/phonid --preset proquint-sha256 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
```

If you tweak the config, re-run preflight to validate and refresh expectations:

```sh
./dist/phonid*/phonid preflight --suggest
```

With preset:

```sh
./dist/phonid*/phonid --preset proquint 1337
```

With config file:

```sh
./dist/phonid*/phonid --config ./public_presets/.proquint.phonidrc.toml 1337
```

Prefixed base input is accepted with `0x` / `0X` (hex), `0b` / `0B` (binary), and `0o` / `0O` (octal):

```sh
./dist/phonid*/phonid 0x539
```

Decode with config file:

```sh
./dist/phonid*/phonid --config ./public_presets/.proquint.phonidrc.toml decode babab-bihun
```

Convert text to SHA-256 hex externally, then encode with the SHA-256-compatible preset:

```sh
HEX=$(python3 -c 'import hashlib; print("0x" + hashlib.sha256(b"hello world").hexdigest())')
./dist/phonid-unix_linux_amd64_v1/phonid  --preset proquint-sha256 $HEX
```

`proquint-tiny` uses canonical ProQuint alphabets (`C=16`, `V=4`) with a single `CVCVC` pattern.

`proquint-sha256` uses 16 `CVCVC` blocks and 15 delimiters to encode exactly 256 bits.

Note: `--preset` is mutually exclusive with `--config`, and is not supported with `preflight`.


4. Decode:

```sh
./dist/phonid*/phonid decode babab-bihun
```

## Public Presets

Phonid includes community-contributed public presets for common use cases. These are available in the [WASM demo](https://iilei.github.io/phonid/) and can be used directly:

```bash
# Use a public preset
curl -o .phonidrc.toml https://raw.githubusercontent.com/iilei/phonid/master/public_presets/.proquint.phonidrc.toml
phonid --config .phonidrc.toml 12345
```

Available presets include:
- **ProQuint** - Standard ProQuint-compatible encoding
- **Tiny** - Minimal character sets for short codes
- **ProQuint SHA-256** - Exact 256-bit reversible encoding space
- **Special** - Unicode-based encoding with special characters

See [public_presets/README.md](public_presets/README.md) for details.

### Contributing a Preset

Want to share your configuration with the community?

```bash
# Quick start
./scripts/create-preset.sh
```

All presets are automatically validated and must:
- Be generated by `phonid preflight --suggest`
- Pass all preflight checks
- Have signed commits
- Serve a distinct use case

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for detailed instructions.

## License

Phonid is released under an open-source license. See the LICENSE file for details.
