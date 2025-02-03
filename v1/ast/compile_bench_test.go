package ast

import (
	"strconv"
	"testing"
)

func BenchmarkRewriteDynamics(b *testing.B) {

	// The choice of query to use is somewhat arbitrary. This query is
	// representative of the ones that result from partial evaluation on IAM
	// data models (e.g., a triple glob match on subject/action/resource.)
	body := MustParseBody(`
		glob.match("a:*", [":"], input.abcdef.x12345);
		glob.match("a:*", [":"], input.abcdef.y12345);
		glob.match("a:*", [":"], input.abcdef.z12345)
	`)
	sizes := []int{1, 10, 100, 1000, 10000, 100000}
	queries := makeQueriesForRewriteDynamicsBenchmark(sizes, body)

	for i := range sizes {
		b.Run(strconv.Itoa(sizes[i]), func(b *testing.B) {
			factory := newEqualityFactory(newLocalVarGenerator("q", nil))
			b.ResetTimer()
			for range b.N {
				for _, body := range queries[i] {
					rewriteDynamics(factory, body)
				}
			}
		})
	}

}

// 4162777	       284.7 ns/op	     272 B/op	       8 allocs/op
// 144060980	     8.3 ns/op	       0 B/op	       0 allocs/op ruleCanHaveRefs
// ...
func BenchmarkResolveNoRefsRule(b *testing.B) {
	rule := MustParseRule(`default allow := false`)
	globals := map[Var]*usedRef{}

	b.ResetTimer()
	for range b.N {
		err := resolveRefsInRule(globals, rule)
		if err != nil {
			b.Fatal(err)
		}
		if len(globals) != 0 {
			b.Fatal("unexpected globals")
		}
	}
}

// 827070	      1397 ns/op	    1152 B/op	      34 allocs/op
// 867757	      1342 ns/op	    1144 B/op	      33 allocs/op	pre-allocate
// 849734	      1349 ns/op	    1096 B/op	      32 allocs/op
func BenchmarkResolveRefsInRegularRule(b *testing.B) {
	rule := MustParseRule(`rule contains x if {
		three := 3
		some y in [1, 2, three]
		x := y
	}`)
	globals := map[Var]*usedRef{}

	b.ResetTimer()
	for range b.N {
		err := resolveRefsInRule(globals, rule)
		if err != nil {
			b.Fatal(err)
		}
		if len(globals) != 0 {
			b.Fatal("unexpected globals")
		}
	}
}

func makeQueriesForRewriteDynamicsBenchmark(sizes []int, body Body) [][]Body {

	queries := make([][]Body, len(sizes))

	for i := range queries {
		queries[i] = make([]Body, sizes[i])
		for j := range sizes[i] {
			queries[i][j] = body.Copy()
		}
	}

	return queries
}
