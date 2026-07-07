package topdown_test

import (
	"cmp"
	"fmt"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"
	inmem "github.com/open-policy-agent/opa/v1/storage/inmem/test"
	"github.com/open-policy-agent/opa/v1/topdown"
)

func BenchmarkBiunifyArrays(b *testing.B) {
	q := topdown.NewQuery(ast.MustParseBody("[1,x,3] = [y,5,6]"))

	ctx := b.Context()

	for b.Loop() {
		if _, err := q.Run(ctx); err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkOrOptionsViaIndexing/unconditional-16                  581562      2065 ns/op    2972 B/op      59 allocs/op
// BenchmarkOrOptionsViaIndexing/function_arg_eq_matching-16       407439      2772 ns/op    4399 B/op      86 allocs/op
// BenchmarkOrOptionsViaIndexing/function_arg_body_eq_matching-16  559994      2202 ns/op    3417 B/op      65 allocs/op
// BenchmarkOrOptionsViaIndexing/rule_input_eq_matching-16         586120      2125 ns/op    2908 B/op      56 allocs/op
// BenchmarkOrOptionsViaIndexing/value_'in'_matching-16            490918      2291 ns/op    3071 B/op      61 allocs/op
// BenchmarkOrOptionsViaIndexing/else_eq_matching-16               556843      2073 ns/op    2908 B/op      56 allocs/op
//
// After 'unconditional' early exit
// --------------------------------
// BenchmarkOrOptionsViaIndexing/unconditional-16                  749073      1587 ns/op    2474 B/op      44 allocs/op
// BenchmarkOrOptionsViaIndexing/function_arg_eq_matching-16       863868      1398 ns/op    2450 B/op      43 allocs/op
func BenchmarkOrOptionsViaIndexing(b *testing.B) {
	tests := []struct {
		note   string
		query  string
		input  string
		module string
	}{
		{
			note:  "unconditional",
			query: `data.p.x = x`,
			input: `{"x": "qux"}`,
			module: `package p
				r := true
				x := r
			`,
		},
		{
			note:  "function arg eq matching",
			query: `data.p.x = x`,
			module: `package p
				f("foo")
				f("bar")
				f("baz")
				f("qux")

				x := f("qux")
			`,
		},
		{
			note:  "function arg body eq matching",
			query: `data.p.x = x`,
			module: `package p
				f(x) if x == "foo"
				f(x) if x == "bar"
				f(x) if x == "baz"
				f(x) if x == "qux"

				x := f("qux")
			`,
		},
		{
			note:  "rule input eq matching",
			query: `data.p.x = x`,
			input: `{"x": "qux"}`,
			module: `package p
				r if input.x == "foo"
				r if input.x == "bar"
				r if input.x == "baz"
				r if input.x == "qux"

				x := r
			`,
		},
		{
			note:  "value 'in' matching",
			query: `data.p.x = x`,
			input: `{"x": "qux"}`,
			module: `package p
				r if input.x in {"foo", "bar", "baz", "qux"}

				x := r
			`,
		},
		{
			note:  "else eq matching",
			query: `data.p.x = x`,
			input: `{"x": "qux"}`,
			module: `package p
				r if {
					input.x == "foo"
				} else if {
					input.x == "bar"
				} else if {
					input.x == "baz"
				} else if {
					input.x == "qux"
				}

				x := r
			`,
		},
	}

	for _, tc := range tests {
		b.Run(tc.note, func(b *testing.B) {
			ctx := b.Context()
			store := inmem.New()
			parsed := ast.MustParseModule(tc.module)

			c := ast.NewCompiler()
			if c.Compile(map[string]*ast.Module{"p": parsed}); c.Failed() {
				b.Fatalf("Unexpected compile error: %v", c.Errors)
			}
			txn := storage.NewTransactionOrDie(ctx, store)
			defer store.Abort(ctx, txn)

			query := topdown.NewQuery(ast.MustParseBody(tc.query)).
				WithCompiler(c).
				WithStore(store).
				WithTransaction(txn).
				WithInput(ast.MustParseTerm(cmp.Or(tc.input, "{}")))

			expIter := func(qr topdown.QueryResult) error {
				if !ast.InternedTerm(true).Equal(qr["x"]) {
					return fmt.Errorf("unexpected result: %v", qr["x"])
				}
				return nil
			}

			for b.Loop() {
				if err := query.Iter(ctx, expIter); err != nil {
					b.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}
