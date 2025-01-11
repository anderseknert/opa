package rego

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/open-policy-agent/opa/internal/runtime"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/open-policy-agent/opa/v1/storage"
	inmem "github.com/open-policy-agent/opa/v1/storage/inmem/test"
	"github.com/open-policy-agent/opa/v1/util/test"
)

func BenchmarkPartialObjectRuleCrossModule(b *testing.B) {
	ctx := context.Background()
	sizes := []int{10, 100, 1000}

	for _, n := range sizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			store := inmem.NewFromObject(map[string]interface{}{})
			mods := test.PartialObjectBenchmarkCrossModule(n)
			query := "data.test.foo"

			input := make(map[string]interface{})
			for idx := range 4 {
				input[fmt.Sprintf("test_input_%d", idx)] = "test_input_10"
			}
			inputAST, err := ast.InterfaceToValue(input)
			if err != nil {
				b.Fatal(err)
			}

			compiler := ast.MustCompileModules(map[string]string{
				"test/foo.rego": mods[0],
				"test/bar.rego": mods[1],
				"test/baz.rego": mods[2],
			})
			info, err := runtime.Term(runtime.Params{})
			if err != nil {
				b.Fatal(err)
			}

			pq, err := New(
				Query(query),
				Compiler(compiler),
				Store(store),
				Runtime(info),
			).PrepareForEval(ctx)

			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for range b.N {
				_, err = pq.Eval(
					ctx,
					EvalParsedInput(inputAST),
					EvalRuleIndexing(true),
					EvalEarlyExit(true),
				)

				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCustomFunctionInHotPath-10    74    16082820 ns/op     13046333 B/op      228399 allocs/op // original
// BenchmarkCustomFunctionInHotPath-10    72    15830449 ns/op     12774346 B/op      222733 allocs/op // custom sprintf
// BenchmarkCustomFunctionInHotPath-10    76    15453395 ns/op     12500588 B/op      211387 allocs/op // slices.SortFunc in indexing
// BenchmarkCustomFunctionInHotPath-10    79    15273517 ns/op     12404279 B/op      205318 allocs/op // sync.Once *set no pointer
// BenchmarkCustomFunctionInHotPath-10    79    14930429 ns/op     12267607 B/op      199646 allocs/op // slices.SortFunc in set sorting
// BenchmarkCustomFunctionInHotPath-10    81	14683766 ns/op	   12098560 B/op	  193556 allocs/op // TypedHashMap
// BenchmarkCustomFunctionInHotPath-10    85	13996864 ns/op	   11553742 B/op	  176555 allocs/op // evalCache IsGround
// BenchmarkCustomFunctionInHotPath-10    86	13975634 ns/op	   11510500 B/op	  171121 allocs/op // reuse evalFunc termslice
// BenchmarkCustomFunctionInHotPath-10    91	13247092 ns/op	   11061042 B/op	  159798 allocs/op // pooled index results
// BenchmarkCustomFunctionInHotPath-10    96	12616755 ns/op	   10924871 B/op	  154131 allocs/op // various fixes
// BenchmarkCustomFunctionInHotPath-10    99	11966397 ns/op	   10832557 B/op	  148439 allocs/op // torin's index fix

func BenchmarkCustomFunctionInHotPath(b *testing.B) {
	ctx := context.Background()
	input := ast.MustParseTerm(mustReadFileAsString(b, "testdata/ast.json"))
	module := ast.MustParseModule(`package test

	r := count(refs)

	refs contains value if {
		walk(input, [_, value])
		is_ref(value)
	}

	is_ref(value) if value.type == "ref"
	is_ref(value) if value[0].type == "ref"`)

	r := New(Query("data.test.r = x"), ParsedModule(module))

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		res, err := pq.Eval(ctx, EvalParsedInput(input.Value))
		if err != nil {
			b.Fatal(err)
		}

		if res == nil {
			b.Fatal("expected result")
		}

		if res[0].Bindings["x"].(json.Number) != "402" {
			b.Fatalf("expected 402, got %v", res[0].Bindings["x"])
		}
	}
}

// Benchmarks of the ACI test data from Regorus
// https://github.com/microsoft/regorus?tab=readme-ov-file#performance

// BenchmarkAciTestBuildAndEval-10    37    30700209 ns/op    16437935 B/op    384211 allocs/op
func BenchmarkAciTestBuildAndEval(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		bundle, err := loader.NewFileLoader().
			WithRegoVersion(ast.RegoV0).
			AsBundle("testdata/aci")
		if err != nil {
			b.Fatal(err)
		}

		input := ast.MustParseTerm(mustReadFileAsString(b, "testdata/aci/input.json"))

		r := New(Query("data.framework.mount_overlay = x"), ParsedBundle("", bundle))

		pq, err := r.PrepareForEval(ctx)
		if err != nil {
			b.Fatal(err)
		}

		res, err := pq.Eval(ctx, EvalParsedInput(input.Value))
		if err != nil {
			b.Fatal(err)
		}

		_ = res
	}
}

// BenchmarkAciTestOnlyEval-10    12752    92188 ns/op    50005 B/op    1062 allocs/op
// BenchmarkAciTestOnlyEval-10    13521	   86647 ns/op	  47448 B/op	 967 allocs/op // ref.CopyNonGround
func BenchmarkAciTestOnlyEval(b *testing.B) {
	ctx := context.Background()

	bundle, err := loader.NewFileLoader().
		WithRegoVersion(ast.RegoV0).
		AsBundle("testdata/aci")
	if err != nil {
		b.Fatal(err)
	}

	input := ast.MustParseTerm(mustReadFileAsString(b, "testdata/aci/input.json"))

	r := New(Query("data.framework.mount_overlay = x"), ParsedBundle("", bundle))

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		res, err := pq.Eval(ctx, EvalParsedInput(input.Value))
		if err != nil {
			b.Fatal(err)
		}
		_ = res
	}
}

// BenchmarkArrayIteration-10
// 15574    77121 ns/op    67249 B/op    1115 allocs/op    // handleErr wrapping, not inlined
// 33862    35864 ns/op     5768 B/op      93 allocs/op    // handleErr only on error, inlined
func BenchmarkArrayIteration(b *testing.B) {
	ctx := context.Background()

	at := make([]*ast.Term, 512)
	for i := range 511 {
		at[i] = ast.StringTerm("a")
	}
	at[511] = ast.StringTerm("v")

	input := ast.NewObject(ast.Item(ast.StringTerm("foo"), ast.ArrayTerm(at...)))
	module := ast.MustParseModule(`package test

	default r := false

	r if input.foo[_] == "v"`)

	r := New(Query("data.test.r = x"), ParsedModule(module))

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		res, err := pq.Eval(ctx, EvalParsedInput(input))
		if err != nil {
			b.Fatal(err)
		}

		if res == nil {
			b.Fatal("expected result")
		}

		if res[0].Bindings["x"].(bool) != true {
			b.Fatalf("expected true, got %v", res[0].Bindings["x"])
		}
	}
}

// BenchmarkSetIteration-10
// 4800    272403 ns/op    80875 B/op    1193 allocs/op    // handleErr wrapping, not inlined
// 4933	   223234 ns/op	   76772 B/op	  681 allocs/op    // handleErr only on error, not inlined
func BenchmarkSetIteration(b *testing.B) {
	ctx := context.Background()

	at := make([]*ast.Term, 512)
	for i := range 512 {
		at[i] = ast.StringTerm(strconv.Itoa(i))
	}

	input := ast.NewObject(ast.Item(ast.StringTerm("foo"), ast.ArrayTerm(at...)))
	module := ast.MustParseModule(`package test

	s := {x | x := input.foo[_]}

	default r := false

	r if s[_] == "not found"`)

	r := New(Query("data.test.r = x"), ParsedModule(module))

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		res, err := pq.Eval(ctx, EvalParsedInput(input))
		if err != nil {
			b.Fatal(err)
		}
		if res == nil {
			b.Fatal("expected result")
		}
		if res[0].Bindings["x"].(bool) != false {
			b.Fatalf("expected false, got %v", res[0].Bindings["x"])
		}
	}
}

// BenchmarkObjectIteration-10
// 12067    99582 ns/op    72830 B/op   1126 allocs/op    // handleErr wrapping, not inlined
// 15358    85080 ns/op    27752 B/op    615 allocs/op    // handleErr only on error, not inlined
func BenchmarkObjectIteration(b *testing.B) {
	ctx := context.Background()

	at := make([][2]*ast.Term, 512)
	for i := range 512 {
		at[i] = ast.Item(ast.StringTerm(strconv.Itoa(i)), ast.StringTerm(strconv.Itoa(i)))
	}

	input := ast.NewObject(ast.Item(ast.StringTerm("foo"), ast.ObjectTerm(at...)))
	module := ast.MustParseModule(`package test

	default r := false

	r if {
		input.foo[_] == "512"
	}
		`)

	r := New(Query("data.test.r = x"), ParsedModule(module))

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		res, err := pq.Eval(ctx, EvalParsedInput(input))
		if err != nil {
			b.Fatal(err)
		}
		if res == nil {
			b.Fatal("expected result")
		}
		if res[0].Bindings["x"].(bool) != false {
			b.Fatalf("expected false, got %v", res[0].Bindings["x"])
		}
	}
}

// Comparing the cost of referencing not found data in Go vs. AST storage
//
// BenchmarkStoreRefNotFound/inmem-go-10         5208    212288 ns/op    160609 B/op     2936 allocs/op
// BenchmarkStoreRefNotFound/inmem-ast-10       13929     90053 ns/op     39614 B/op     1012 allocs/op
func BenchmarkStoreRefNotFound(b *testing.B) {
	ctx := context.Background()

	things := make(map[string]map[string]string, 100)
	for i := range 100 {
		things[strconv.Itoa(i)] = map[string]string{"foo": "bar"}
	}

	stores := map[string]storage.Store{
		"inmem-go":  inmem.NewFromObject(map[string]any{"things": things}),
		"inmem-ast": inmem.NewFromObjectWithASTRead(map[string]any{"things": things}),
	}
	policy := `package p

r contains true if {
	data.things[_].bar
}
`
	for name, store := range stores {
		b.Run(name, func(b *testing.B) {
			r := New(
				Query("data.p.r = x"),
				Store(store),
				ParsedModule(ast.MustParseModule(policy)),
				GenerateJSON(func(*ast.Term, *EvalContext) (any, error) {
					return nil, nil
				}),
			)

			pq, err := r.PrepareForEval(ctx)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				res, err := pq.Eval(ctx)
				if err != nil {
					b.Fatal(err)
				}

				_ = res
			}
		})
	}
}

// 242.5 ns/op     168 B/op      7 allocs/op  // original implementation
// 176.7 ns/op      96 B/op      4 allocs/op  // sync.Pool in ptr.ValuePtr (saving 1 alloc/op per path part)
func BenchmarkStoreRead(b *testing.B) {
	ctx := context.Background()
	store := inmem.NewFromObjectWithASTRead(map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": "qux",
			},
		},
	})

	txn, err := store.NewTransaction(ctx)
	if err != nil {
		b.Fatal(err)
	}

	ref := ast.MustParseRef("data.foo.bar.baz")

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		// 1 alloc/op
		path, err := storage.NewPathForRef(ref)
		if err != nil {
			b.Fatal(err)
		}

		// 3 allocs/op (down from 6)
		// turns each string in path into a StringTerm only to use it
		// for a Get call in storage (ptr.ValuePtr)
		v, err := store.Read(ctx, txn, path)
		if err != nil {
			b.Fatal(err)
		}

		if v == nil {
			b.Fatal("expected value")
		}
	}
}

// 233337	      5730 ns/op	    5737 B/op	      93 allocs/op
// 229280	      5222 ns/op	    5639 B/op	      89 allocs/op // ref.CopyNonGround
func BenchmarkTrivialPolicy(b *testing.B) {
	ctx := context.Background()
	r := New(
		ParsedQuery(ast.MustParseBody("data.p.r = x")),
		ParsedModule(ast.MustParseModule(`package p
		r := 1`)),
		GenerateJSON(func(*ast.Term, *EvalContext) (any, error) {
			return nil, nil
		}),
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for range b.N {
		_, err := pq.Eval(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func mustReadFileAsString(b *testing.B, path string) string {
	b.Helper()

	bs, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}

	return string(bs)
}
