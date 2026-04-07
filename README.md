# Phonid

> **Try it live:** Experiment with different configurations and encoding patterns in the [interactive playground](https://iilei.github.io/phonid/)

Phonid is a Go library for encoding and decoding numeric identifiers into **pronounceable, human-friendly fantasy words** while preserving **strict mathematical reversibility**.

This project is inspired by **Proquints** and specifically by the original publication, *Proquints: Identifiers that are Readable, Spellable, and Pronounceable*:
https://arxiv.org/html/0901.4016
<!-- archived: https://web.archive.org/web/20260401000408/https://arxiv.org/html/0901.4016 -->

Phonid generalizes that idea into a configurable, extensible system that allows different "phonetic languages" (e.g. Minion-like, Elvish-like) **without ever sacrificing bidirectional decodability**.

## Design Goals

Phonid is built around a small number of hard design constraints:

* Every encoded identifier **must be uniquely decodable** (given the configuration and seed).
* No configuration option may introduce lossy or ambiguous transformations.
* All decoding decisions must be **deterministic and non-heuristic**.
* Performance should be predictable and close to constant time.

The result is a system that is playful in output, but intentionally conservative in its formal model.

## Core Encoding Model

Phonid represents numbers as **words composed of consonants (C) and vowels (V)** according to explicitly defined word templates.

### 2. Explicit Word Templates

For each pattern, the configuration defines a **finite set of word templates**, such as:

* Length 3: `CVC`, `CVV`
* Length 5: `CVCCV`, `CVCVC`

A template is a positional blueprint that determines:

* Which alphabet (consonant or vowel) is used at each position
* How many symbols are available per position
* How bits are packed and unpacked

Templates are not inferred or generated implicitly — only explicitly declared templates exist.

### 3. Mandatory Template Disjointness

To guarantee unambiguous decoding, **all templates must be disjoint**:

* No template may be a prefix of another template
* No shorter template may appear as a contiguous substring of a longer template
* Templates are validated at configuration load time

Because of this rule, **template recognition is trivial and deterministic**:

1. Determine the word length
2. Derive the C/V signature of the word
3. Perform an exact lookup of `(length, template)`

No backtracking, greedy matching, or heuristics are ever required.

## Bidirectional Safety Guarantee

Under the above constraints, Phonid guarantees:

* A bijective mapping between numbers and phonetic words
* Lossless decoding for all valid inputs
* Stable behavior across versions (subject to semantic versioning rules)

Formally:

> If word lengths are prime, templates are explicit, and templates are pairwise disjoint, then the mapping between numeric space and phonetic space is uniquely decodable.

## Configuration Philosophy

Phonid configurations are intentionally constrained.

The configuration language allows:

* Defining alphabets (consonants, vowels)
* Declaring valid word templates per word length
* Selecting enabled word lengths

The configuration **does not allow**:

* Context-sensitive rules
* Any transformation that cannot be mathematically reversed


## Versioning and Stability

Phonid follows Semantic Versioning (SemVer).

.. warning::

While the major version is `0.x.y`, **breaking changes may occur at any time**.
Stability guarantees apply only after `v1.0.0`.

## Local Dev

Install GoReleaser

```sh
go install github.com/goreleaser/goreleaser/v2@latest
```

### Build
```sh
goreleaser build --snapshot --clean --single-target
```

### Generate preflight checks

```sh
./dist/phonid*/phonid preflight --suggest >| .phonidrc.tmp && mv .phonidrc.tmp .phonidrc.toml
```

You may tweak the config as you like and check the results using

```sh
./dist/phonid*/phonid preflight --suggest
```

Once you are happy with the resulting ids, persist the preflight-expectation table as follows:

```sh
./dist/phonid*/phonid preflight --suggest  >| .phonidrc.tmp && mv .phonidrc.tmp .phonidrc.toml
```

### Encode any number

```sh
./dist/phonid*/phonid 4711
```

Hex input is accepted automatically when prefixed with `0x` or `0X`:

```sh
./dist/phonid*/phonid 0x539
```

With explicit config:

```sh
 ./dist/phonid*/phonid  --config ./public_presets/.proquint.phonidrc.toml 1337

```

With built-in preset:

```sh
./dist/phonid*/phonid --preset proquint 1337
```

Yields: `babab-bihun`

### Decode any phonid string

```sh
./dist/phonid*/phonid decode babab-bihun
```

Yields: `1337`

With explicit config:

```sh
./dist/phonid*/phonid --config ./public_presets/.proquint.phonidrc.toml decode babab-bihun
```

### Generate a one-way digest

Digest output is intentionally non-reversible and may collide.

If input is prefixed with 0x or 0X, it is parsed as a hexadecimal integer and reduced modulo the active capacity.

```sh
./dist/phonid*/phonid digest 0x10FA9EB
```

Without `0x` prefix, input is treated as text and hashed with SHA-256 before modulo reduction:

```sh
./dist/phonid*/phonid digest "hello world"
```

With tiny built-in preset:

```sh
./dist/phonid*/phonid --preset proquint-tiny digest "hello world"
```

`proquint-tiny` uses canonical ProQuint alphabets (`C=16`, `V=4`) with a single `CVCVC` pattern.

Note: `--preset` is mutually exclusive with `--config`, and is not supported with `preflight`.

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
