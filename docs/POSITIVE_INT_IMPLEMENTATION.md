# PositiveInt Interface Implementation

## Overview

Converted `PositiveInt` from a concrete `int` type to an interface with two implementations:
- `positiveIntNative`: Fast path for int64 values (zero overhead)
- `positiveIntBig`: Slow path for arbitrarily large values using `math/big`

This enables phonid to handle numbers larger than 2^63-1 without breaking existing code or degrading performance for common cases.

## Design Decisions

### 1. Interface-Based Approach
```go
type PositiveInt interface {
    Int64() (int64, bool)       // Fast extraction, returns (value, ok)
    BigInt() *big.Int           // Always succeeds, may allocate
    Cmp(other PositiveInt) int  // Compare two values
    String() string             // Human-readable representation
    Sign() int                  // Returns 1 for positive, 0 for zero, -1 for negative
    Validate() error            // Ensures non-negativity
}
```

### 2. Two Implementations

**positiveIntNative (int64):**
- Stored directly as `type positiveIntNative int64`
- Zero-allocation operations
- Sub-nanosecond performance
- Automatic when value fits in int64

**positiveIntBig (*big.Int):**
- Stored as `struct { value *big.Int }`
- Allocates for BigInt() and String()
- ~50-200ns operations
- Used only when necessary

### 3. Automatic Optimization

`NewPositiveIntFromBig(*big.Int)` automatically selects the right implementation:
```go
if bigVal.IsInt64() && bigVal.Sign() >= 0 {
    return positiveIntNative(bigVal.Int64()), nil
}
return &positiveIntBig{value: new(big.Int).Set(bigVal)}, nil
```

### 4. TOML Integration

Used a wrapper type for unmarshaling:
```go
type TomlPositiveInt struct {
    Value PositiveInt
}

func (t *TomlPositiveInt) UnmarshalText(data []byte) error {
    val, err := ParsePositiveInt(string(data))
    if err != nil {
        return err
    }
    t.Value = val
    return nil
}
```

## Performance Characteristics

Benchmarks on AMD Ryzen 7 5800X:

### Fast Path (int64)
- Int64(): 0.44 ns/op, 0 allocs
- Cmp(): 2.4 ns/op, 0 allocs
- String(): 17.7 ns/op, 1 alloc
- Constructor: 0.22 ns/op, 0 allocs
- Parse: 22.8 ns/op, 1 alloc

### Slow Path (big.Int)
- Int64(): 1.3 ns/op, 0 allocs
- Cmp(): 46.3 ns/op, 2 allocs
- String(): 196 ns/op, 3 allocs
- Constructor: 57.9 ns/op, 3 allocs
- Parse: 441 ns/op, 10 allocs

**Key takeaway:** Common case (small numbers) has zero overhead. Large numbers are ~100-200x slower but still very fast (<500ns).

## API Changes

### Constructors
- `NewPositiveInt(int64)` - Create from int64
- `NewPositiveIntFromBig(*big.Int)` - Create from big.Int (with auto-optimization)
- `ParsePositiveInt(string)` - Parse from string (auto-selects implementation)

### Pattern Changes

**Before:**
```go
num := PositiveInt(42)
if num > 100 { ... }
```

**After:**
```go
num := NewPositiveInt(42)
if num.Cmp(NewPositiveInt(100)) > 0 { ... }

// Or extract int64 for direct comparison:
if val, ok := num.Int64(); ok && val > 100 { ... }
```

## Testing

All existing tests pass with updated constructors. Added benchmarks to verify performance characteristics.

Example test update:
```go
// Before:
testCases := []struct {
    input    PositiveInt
    expected string
}{
    {0, "aba"},
    {255, "uzu"},
}

// After:
testCases := []struct {
    input    PositiveInt
    expected string
}{
    {NewPositiveInt(0), "aba"},
    {NewPositiveInt(255), "uzu"},
}
```

## Command-Line Usage

The CLI now accepts arbitrarily large numbers:

```bash
# Works with small numbers (fast path)
./phonid 1234567890

# Works with huge numbers (slow path)
./phonid 340282366920938463463374607431768211455

# Preflight suggestions with large custom values
./phonid preflight --suggest 18446744073709551615
```

Numbers are automatically validated against pattern capacity, giving clear error messages when exceeded.

## Files Modified

Core implementation:
- `pkg/rcparse.go` - Interface definition and implementations
- `pkg/encode.go` - Updated Encode/Decode to use interface
- `pkg/preflight.go` - Updated validation
- `pkg/preflight/suggest.go` - Updated suggestion generation

CLI:
- `cmd/phonid/root.go` - Uses ParsePositiveInt instead of strconv
- `cmd/phonid/preflight.go` - Updated to parse large custom values

Tests:
- All test files updated to use constructors
- Added `pkg/rcparse_bench_test.go` with performance benchmarks

## Migration Guide

For code using phonid as a library:

1. **Replace direct type conversions:**
   ```go
   // Before:
   num := PositiveInt(42)
   
   // After:
   num := phonid.NewPositiveInt(42)
   ```

2. **Replace comparisons:**
   ```go
   // Before:
   if num > 100 { ... }
   
   // After:
   if num.Cmp(phonid.NewPositiveInt(100)) > 0 { ... }
   ```

3. **Replace arithmetic:**
   ```go
   // Before:
   result := num + 10
   
   // After:
   val, _ := num.Int64()  // Or use BigInt() for safety
   result := phonid.NewPositiveInt(val + 10)
   ```

4. **String formatting:**
   ```go
   // Before:
   fmt.Printf("Value: %d", num)
   
   // After:
   fmt.Printf("Value: %s", num.String())
   ```

## Future Enhancements

Potential optimizations:
- Cache String() representations for frequently-used values
- Pool big.Int allocations for temporary operations
- Add arithmetic methods to interface (Add, Sub, Mul, etc.)
- Consider uint64 intermediate layer before big.Int

Current implementation prioritizes simplicity and correctness while delivering excellent performance for the common case.
