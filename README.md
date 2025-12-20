# Phonid

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

### 1. Prime-Length Words

Each encoded word has a length that is a **prime number**:

* Required: `3`, `5`
* Optional: `7`, `11`

Using prime lengths prevents accidental re-segmentation and guarantees that words are treated as atomic units during decoding.

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

* Variable-length inference
* Optional or greedy template matching
* Context-sensitive rules
* Any transformation that cannot be mathematically reversed

There is no "unsafe" mode.

## Performance Characteristics

Encoding and decoding operate in predictable time:

* Word analysis is linear in word length (max 11 symbols)
* Template resolution is constant-time lookup
* Bit packing and unpacking are table-driven

In practice, all operations are effectively constant time.

## Versioning and Stability

Phonid follows Semantic Versioning (SemVer).

.. warning::

While the major version is `0.x.y`, **breaking changes may occur at any time**.
Stability guarantees apply only after `v1.0.0`.

## License

Phonid is released under an open-source license. See the LICENSE file for details.
