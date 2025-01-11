package topdown

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
)

func BenchmarkBulkStartsWithNaive(b *testing.B) {
	data := generateBulkStartsWithInput()
	ctx := context.Background()
	store := inmem.NewFromObject(data)

	compiler := ast.MustCompileModules(map[string]string{
		"test.rego": `
package test

result if {
  startswith(data.strings[_], data.prefixes[_])
}
`,
	})

	query, err := compiler.QueryCompiler().Compile(ast.MustParseBody("data.test.result"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		err := storage.Txn(ctx, store, storage.TransactionParams{}, func(txn storage.Transaction) error {

			q := NewQuery(query).
				WithCompiler(compiler).
				WithStore(store).
				WithTransaction(txn)

			_, err := q.Run(ctx)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBulkStartsWithOptimized(b *testing.B) {
	data := generateBulkStartsWithInput()
	ctx := context.Background()
	store := inmem.NewFromObject(data)

	compiler := ast.MustCompileModules(map[string]string{
		"test.rego": `
package test

result if {
  strings.any_prefix_match(data.strings, data.prefixes)
}
`,
	})

	query, err := compiler.QueryCompiler().Compile(ast.MustParseBody("data.test.result"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		err := storage.Txn(ctx, store, storage.TransactionParams{}, func(txn storage.Transaction) error {

			q := NewQuery(query).
				WithCompiler(compiler).
				WithStore(store).
				WithTransaction(txn)

			_, err := q.Run(ctx)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func generateBulkStartsWithInput() map[string]interface{} {
	var strs, prefixes []string
	for i := 0; i < 1000; i++ {
		strs = append(strs, fmt.Sprintf("aabbccddeeffgghhiijjkkllmmnnoopp_%d", i))
	}
	for i := 0; i < 100; i++ {
		prefixes = append(prefixes, fmt.Sprintf("aabbccddeeffgghhiijjkkllmmnnoorr_%d", i))
	}
	return map[string]interface{}{
		"strings":  strs,
		"prefixes": prefixes,
	}
}

func BenchmarkSplit(b *testing.B) {
	bctx := BuiltinContext{}
	operands := []*ast.Term{
		ast.StringTerm("a.b.c.d.e"),
		ast.StringTerm("."),
	}

	exp := eqIter(ast.ArrayTerm(
		ast.StringTerm("a"),
		ast.StringTerm("b"),
		ast.StringTerm("c"),
		ast.StringTerm("d"),
		ast.StringTerm("e"),
	))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := builtinSplit(bctx, operands, exp); err != nil {
			b.Fatal(err)
		}
	}
}

// Now down to 2 allocations per iteration for ASCII strings, more for non-ASCII as that requires
// string/rune conversion. 2 allocations unavoidable - 1 for the new Term and 1 for its Value.
func BenchmarkSubstring(b *testing.B) {
	operands := []*ast.Term{
		// insert any non-asci character to see the difference of that optimization
		ast.StringTerm("The quick brown fox jumps over the lazy dog"),
		ast.InternedIntNumberTerm(6),
		ast.InternedIntNumberTerm(10),
	}

	iter := eqIter(ast.StringTerm("ick brown "))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := builtinSubstring(BuiltinContext{}, operands, iter); err != nil {
			b.Fatal(err)
		}
	}
}

// Unicode
// BenchmarkIndexOf-10    	10498884	       114.0 ns/op	     176 B/op	       1 allocs/op
//
// ASCII
// BenchmarkIndexOf-10    	36625468	        31.57 ns/op	       0 B/op	       0 allocs/op
func BenchmarkIndexOf(b *testing.B) {
	operands := []*ast.Term{
		ast.StringTerm("The quick brown fox jumps over the lazy dog"),
		ast.StringTerm("dog"),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := builtinIndexOf(BuiltinContext{}, operands, eqIter(ast.InternedIntNumberTerm(40))); err != nil {
			b.Fatal(err)
		}
	}
}

func eqIter(a *ast.Term) func(*ast.Term) error {
	return func(b *ast.Term) error {
		if !a.Equal(b) {
			return fmt.Errorf("expected %v equal to %v", a, b)
		}
		return nil
	}
}
