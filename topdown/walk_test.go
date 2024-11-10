package topdown

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
)

func BenchmarkWalkBuiltin(b *testing.B) {
	trm := ast.MustParseTerm(`{
		"foo": [1, 2, 3],
		"bar": [4, 5, 6],
		"baz": {
			"qux": [7, 8, 9]
		},
		"quz": {
			"qux": [10, 11, {
				"corge": [12, 13, 14]
			}]
		},
		"a": [1, 2, 3],
		"b": [4, 5, 6],
		"c": {
			"d": [7, 8, 9]
		},
		"e": {
			"f": [10, 11, {
				"g": [12, 13, 14]
			}]
		},
		"h": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"],
	}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := walk(nil, nil, trm, func(*ast.Term) error {
			return nil
		}); err != nil {
			b.Fatal(err)
		}
	}
}
