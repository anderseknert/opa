// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/util"
)

// BenchmarkParseModuleRulesBase gives a baseline for parsing modules with
// what are extremely simple rules.
func BenchmarkParseModuleRulesBase(b *testing.B) {
	sizes := []int{1, 10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			mod := generateModule(size)
			runParseModuleBenchmark(b, mod)
		})
	}
}

// BenchmarkParseStatementBasic gives a baseline for parsing a simple
// statement with a single call and two variables
func BenchmarkParseStatementBasicCall(b *testing.B) {
	runParseStatementBenchmark(b, `a+b`)
}

func BenchmarkParseStatementMixedJSON(b *testing.B) {
	// While nothing in OPA is Kubernetes specific, the webhook admission
	// request payload makes for an interesting parse test being a moderately
	// deep nested object with several different types of values.
	stmt := `{"uid":"d8fdc6db-44e1-11e9-a10f-021ca99d149a","kind":{"group":"apps","version":"v1beta1","kind":"Deployment"},"resource":{"group":"apps","version":"v1beta1","resource":"deployments"},"namespace":"opa-test","operation":"CREATE","userInfo":{"username":"user@acme.com","groups":["system:authenticated"]},"object":{"metadata":{"name":"nginx","namespace":"torin-opa-test","uid":"d8fdc047-44e1-11e9-a10f-021ca99d149a","generation":1,"creationTimestamp":"2019-03-12T16:14:01Z","labels":{"run":"nginx"}},"spec":{"replicas":1,"selector":{"matchLabels":{"run":"nginx"}},"template":{"metadata":{"creationTimestamp":null,"labels":{"run":"nginx"}},"spec":{"containers":[{"name":"nginx","image":"nginx","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","imagePullPolicy":"Always"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","securityContext":{},"schedulerName":"default-scheduler"}},"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxUnavailable":"25%","maxSurge":"25%"}},"revisionHistoryLimit":2,"progressDeadlineSeconds":600},"status":{}},"oldObject":null}`
	runParseStatementBenchmark(b, stmt)
}

// BenchmarkParseStatementSimpleArray gives a baseline for parsing arrays of strings.
// There is no nesting, so all test cases are flat array structures.
func BenchmarkParseStatementSimpleArray(b *testing.B) {
	sizes := []int{1, 10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			stmt := generateArrayStatement(size)
			runParseStatementBenchmark(b, stmt)
		})
	}
}

func TestParseStatementSimpleArray(b *testing.T) {
	sizes := []int{10} // , 10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprint(size), func(b *testing.T) {
			stmt := generateArrayStatement(size)
			_, err := ParseStatement(stmt)
			if err != nil {
				b.Fatalf("Unexpected error: %s", err)
			}
		})
	}
}

// BenchmarkParseStatementNestedObjects gives a baseline for parsing objects.
// This includes "flat" ones and more deeply nested varieties.
func BenchmarkParseStatementNestedObjects(b *testing.B) {
	sizes := [][]int{{1, 1}, {5, 1}, {10, 1}, {1, 5}, {1, 10}, {5, 5}} // Note: 10x10 will essentially hang while parsing
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(b *testing.B) {
			stmt := generateObjectStatement(size[0], size[1])
			runParseStatementBenchmark(b, stmt)
		})
	}
}

func BenchmarkParseStatementNestedObjectsOrSets(b *testing.B) {
	sizes := []int{1, 5, 10, 15, 20}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			stmt := generateObjectOrSetStatement(size)
			runParseStatementBenchmarkWithError(b, stmt)
		})
	}
}

func BenchmarkParseBasicABACModule(b *testing.B) {
	mod := `
	package app.abac

	default allow = false

	allow if {
		user_is_owner
	}

	allow if {
		user_is_employee
		action_is_read
	}

	allow if {
		user_is_employee
		user_is_senior
		action_is_update
	}

	allow if {
		user_is_customer
		action_is_read
		not pet_is_adopted
	}

	user_is_owner if {
		data.user_attributes[input.user].title == "owner"
	}

	user_is_employee if {
		data.user_attributes[input.user].title == "employee"
	}

	user_is_customer if {
		data.user_attributes[input.user].title == "customer"
	}

	user_is_senior if {
		data.user_attributes[input.user].tenure > 8
	}

	action_is_read if {
		input.action == "read"
	}

	action_is_update if {
		input.action == "update"
	}

	pet_is_adopted if {
		data.pet_attributes[input.resource].adopted == true
	}
	`
	runParseModuleBenchmark(b, mod)
}

func runParseModuleBenchmark(b *testing.B, mod string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseModuleWithOpts("", mod, ParserOptions{AllFutureKeywords: true})
		if err != nil {
			b.Fatalf("Unexpected error: %s", err)
		}
	}
}

func runParseStatementBenchmark(b *testing.B, stmt string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseStatement(stmt)
		if err != nil {
			b.Fatalf("Unexpected error: %s", err)
		}
	}
}

func runParseStatementBenchmarkWithError(b *testing.B, stmt string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseStatement(stmt)
		if err == nil {
			b.Fatalf("Expected error: %s", err)
		}
	}
}

func generateModule(numRules int) string {
	mod := "package bench\n"
	for i := 0; i < numRules; i++ {
		mod += fmt.Sprintf("p%d if { input.x%d = %d }\n", i, i, i)
	}
	return mod
}

func generateArrayStatement(size int) string {
	a := make([]string, size)
	for i := 0; i < size; i++ {
		a[i] = fmt.Sprintf("entry-%d", i)
	}
	return string(util.MustMarshalJSON(a))
}

func generateObjectStatement(width, depth int) string {
	o := generateObject(width, depth)
	return string(util.MustMarshalJSON(o))
}

func generateObject(width, depth int) map[string]interface{} {
	o := map[string]interface{}{}
	for i := 0; i < width; i++ {
		key := fmt.Sprintf("entry-%d", i)
		if depth <= 1 {
			o[key] = "value"
		} else {
			o[key] = generateObject(width, depth-1)
		}
	}
	return o
}

func generateObjectOrSetStatement(depth int) string {
	s := strings.Builder{}
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&s, `{a%d:a%d|`, i, i)
	}
	return s.String()
}

// BenchmarkParseAnnotations/with_annotations-10         	   34660	     39087 ns/op	   32705 B/op	     327 allocs/op
// BenchmarkParseAnnotations/without_annotations-10      	   63753	     18876 ns/op	   14195 B/op	     157 allocs/op

// BenchmarkParseAnnotations/with_annotations-10         	   47551	     25211 ns/op	   29599 B/op	     296 allocs/op
// BenchmarkParseAnnotations/without_annotations-10      	  119618	      9952 ns/op	   11336 B/op	     135 allocs/op

func BenchmarkParseAnnotations(b *testing.B) {
	policy := `# METADATA
# title: pkg
# description: a package
package pkg

# METADATA
# title: rule
# description: a rule
# scope: document
rule.foo.bar := true
`
	var module *Module

	cases := []struct {
		note               string
		processAnnotations bool
		expectAnnotations  int
	}{
		{"with annotations", true, 2},
		{"without annotations", false, 0},
	}

	capabilities := CapabilitiesForThisVersion()

	for _, tc := range cases {
		b.Run(tc.note, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				module = MustParseModuleWithOpts(policy, ParserOptions{
					ProcessAnnotation: tc.processAnnotations,
					Capabilities:      capabilities,
				})
			}

			if len(module.Annotations) != tc.expectAnnotations {
				b.Fatalf("Expected %d annotations but got %d", tc.expectAnnotations, len(module.Annotations))
			}
		})
	}
}

// Without strings.Contains check for escape sequences
// BenchmarkParseString-10    	 3838826	       314.3 ns/op	     368 B/op	       7 allocs/op
//
// With strings.Contains check for escape sequences
// BenchmarkParseString-10    	12074408	        96.11 ns/op	     176 B/op	       3 allocs/op
//
// Remaining 3 allocations are:
// 1. The Term pointer
// 2. The string to Value interface conversion
// 3. Loc() call creating a copy of the location
//
// Neither of these can easily be avoided, but the string to Value interface conversion could
// be avoided if we interned known strings in Value form, or allowed the caller to pass a map
// of known strings to the parser.
func BenchmarkParseString(b *testing.B) {
	p := Parser{
		s: &state{
			lit: `"hello world"`,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := p.parseString()
		if t == nil {
			b.Fatal("Expected string token")
		}
	}
}
