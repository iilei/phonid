# Phonid

> **Try it live:** Experiment with different configurations and encoding patterns in the [interactive playground](https://iilei.github.io/phonid/)

Phonid is a Go library for encoding and decoding numeric identifiers into **pronounceable, human-friendly fantasy words** while preserving **strict mathematical reversibility**.

Inspired by ProQuint, Phonid generalizes the idea into a configurable, extensible system that allows different "phonetic languages" (e.g. Minion-like, Elvish-like) **without ever sacrificing bidirectional decodability**.

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

For each allowed word length, the configuration defines a **finite set of word templates**, such as:

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

## Seeded Permutations

Phonid supports **seeded permutations** of symbol order.

A seed is used to deterministically shuffle consonant and vowel alphabets before encoding. This produces project-specific phonetic sequences while preserving reversibility.

* Different seeds yield different word sequences for the same number
* The numeric space itself is never altered
* No collisions or gaps are introduced

Without knowledge of the seed, the numeric meaning of a word is non-obvious, while decoding remains exact for authorized users.

## Configuration Philosophy

Phonid configurations are intentionally constrained.

The configuration language allows:

* Defining alphabets (consonants, vowels)
* Declaring valid word templates per word length
* Selecting enabled word lengths
* Choosing a permutation seed

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

With explicit config:

```sh
./dist/phonid*/phonid 4711 --config .phonidrc.toml
```

## License

Phonid is released under an open-source license. See the LICENSE file for details.
