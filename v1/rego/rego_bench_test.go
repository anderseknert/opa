package rego

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/open-policy-agent/opa/internal/runtime"
	"github.com/open-policy-agent/opa/v1/ast"
	inmem "github.com/open-policy-agent/opa/v1/storage/inmem/test"
	"github.com/open-policy-agent/opa/v1/util/test"
)

func BenchmarkPartialObjectRuleCrossModule(b *testing.B) {
	ctx := context.Background()
	sizes := []int{10, 100, 1000}

	for _, n := range sizes {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			store := inmem.NewFromObject(map[string]interface{}{})
			mods := test.PartialObjectBenchmarkCrossModule(n)
			query := "data.test.foo"

			input := make(map[string]interface{})
			for idx := 0; idx <= 3; idx++ {
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

			for i := 0; i < b.N; i++ {
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

	bs, err := os.ReadFile("testdata/ast.json")
	if err != nil {
		b.Fatal(err)
	}

	input := ast.MustParseTerm(string(bs))
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

	for i := 0; i < b.N; i++ {
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
