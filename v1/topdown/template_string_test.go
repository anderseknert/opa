package topdown

import (
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

var tests = []struct {
	note   string
	parts  *ast.Array
	expRes *ast.Term
	expErr string
}{
	{
		note:   "no parts",
		parts:  ast.NewArray(),
		expRes: ast.StringTerm(""),
	},
	{
		note:   "single number interned",
		parts:  ast.NewArray(ast.NumberTerm("10")),
		expRes: ast.StringTerm("10"),
	},
	{
		note:   "single string part interned term",
		parts:  ast.NewArray(ast.StringTerm("scope")),
		expRes: ast.StringTerm("scope"),
	},
	{
		note:   "single string part",
		parts:  ast.NewArray(ast.StringTerm("foo")),
		expRes: ast.StringTerm("foo"),
	},
	{
		note:   "single undefined part",
		parts:  ast.NewArray(ast.SetTerm()),
		expRes: ast.StringTerm("<undefined>"),
	},
	{
		note:   "primitives",
		parts:  ast.NewArray(ast.StringTerm("foo"), ast.NumberTerm("42"), ast.BooleanTerm(false), ast.NullTerm()),
		expRes: ast.StringTerm("foo42falsenull"),
	},
	{
		note: "collections",
		parts: ast.NewArray(
			ast.SetTerm(ast.ArrayTerm()), ast.StringTerm(" "),
			ast.SetTerm(ast.ArrayTerm(ast.StringTerm("a"), ast.StringTerm("b"))), ast.StringTerm(" "),
			ast.SetTerm(ast.SetTerm()), ast.StringTerm(" "),
			ast.SetTerm(ast.SetTerm(ast.StringTerm("c"))), ast.StringTerm(" "),
			ast.SetTerm(ast.ObjectTerm()), ast.StringTerm(" "),
			ast.SetTerm(ast.ObjectTerm(ast.Item(ast.StringTerm("d"), ast.StringTerm("e")))),
		),
		expRes: ast.StringTerm(`[] ["a", "b"] set() {"c"} {} {"d": "e"}`),
	},
	{
		note:   "multiple outputs",
		parts:  ast.NewArray(ast.SetTerm(ast.BooleanTerm(true), ast.BooleanTerm(false))),
		expErr: "eval_conflict_error: template-strings must not produce multiple outputs",
	},
}

func TestBuiltinTemplateString(t *testing.T) {
	for _, tc := range tests {
		t.Run(tc.note, func(t *testing.T) {
			var result *ast.Term

			bctx := BuiltinContext{}
			err := builtinTemplateString(bctx, []*ast.Term{ast.NewTerm(tc.parts)}, func(t *ast.Term) error {
				result = t
				return nil
			})

			if tc.expErr == "" {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if ast.Compare(tc.expRes, result) != 0 {
					t.Fatalf("Expected result:\n\n%s\n\ngot:\n\n%s", tc.expRes, result)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				if act := err.Error(); !strings.Contains(act, tc.expErr) {
					t.Fatalf("Expected error to contain:\n\n%s\n\ngot:\n\n%s", tc.expErr, act)
				}
			}
		})
	}
}

// no_parts-16                                  40689573	        29.66 ns/op	       0 B/op	       0 allocs/op
// single_string_part_interned_term-16         	31388037	        37.91 ns/op	       0 B/op	       0 allocs/op
// single_string_part-16                       	18682736	        64.31 ns/op	      43 B/op	       3 allocs/op
// single_undefined_part-16                    	18175942	        66.25 ns/op	      56 B/op	       3 allocs/op
// primitives-16                               	13213772	        89.80 ns/op	      56 B/op	       3 allocs/op
// single_number_interned-16                   	34081790	        35.68 ns/op	       0 B/op	       0 allocs/op
// collections-16                              	 3531015	       338.4 ns/op	     232 B/op	       9 allocs/op
// multiple_outputs-16                         	14427541	        82.97 ns/op	     104 B/op	       3 allocs/op
func BenchmarkBuiltinTemplateString(b *testing.B) {
	for _, tc := range tests {
		b.Run(tc.note, func(b *testing.B) {
			oper := []*ast.Term{ast.NewTerm(tc.parts)}
			iter := eqIter(tc.expRes)

			for b.Loop() {
				_ = builtinTemplateString(BuiltinContext{}, oper, iter)
			}
		})
	}
}
