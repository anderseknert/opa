// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"strconv"
	"strings"
	"unique"

	"github.com/open-policy-agent/opa/v1/ast"
)

type undo struct {
	k *ast.Term
	u *bindings
}

func (u *undo) Undo() {
	if u == nil || u.u == nil {
		// Allow call on zero value of Undo for ease-of-use.
		// Call on empty unifier undos a no-op unify operation.
		return
	}
	delete(u.u.values, unique.Make(u.k.Value.(ast.Var)))
}

type bindings struct {
	id     uint64
	values map[unique.Handle[ast.Var]]bindingKeyValue
	instr  *Instrumentation
}

type bindingKeyValue struct {
	key, val *ast.Term
	bindings *bindings
}

func newBindings(id uint64, instr *Instrumentation) *bindings {
	return &bindings{id: id, instr: instr}
}

// newBindingsWithSize creates bindings pre-sized for the expected number of entries.
// This avoids over-allocation when the binding count is known in advance (e.g., function arguments).
// For sizeHint <= maxLinearScan, it uses array mode; for larger hints, it pre-allocates a map.
func newBindingsWithSize(id uint64, instr *Instrumentation, sizeHint int) *bindings {
	return &bindings{id: id, instr: instr}
}

func (u *bindings) Iter(caller *bindings, iter func(*ast.Term, *ast.Term) error) error {
	for _, v := range u.values {
		if err := iter(v.key, u.PlugNamespaced(v.key, caller)); err != nil {
			return err
		}
	}
	return nil
}

func (u *bindings) Namespace(x ast.Node, caller *bindings) {
	vis := namespacingVisitor{
		b:      u,
		caller: caller,
	}
	ast.NewGenericVisitor(vis.Visit).Walk(x)
}

func (u *bindings) Plug(a *ast.Term) *ast.Term {
	return u.PlugNamespaced(a, nil)
}

func (u *bindings) PlugNamespaced(a *ast.Term, caller *bindings) *ast.Term {
	if u != nil && u.instr != nil {
		u.instr.startTimer(evalOpPlug)
		t := u.plugNamespaced(a, caller)
		u.instr.stopTimer(evalOpPlug)
		return t
	}

	return u.plugNamespaced(a, caller)
}

func (u *bindings) plugNamespaced(a *ast.Term, caller *bindings) *ast.Term {
	switch v := a.Value.(type) {
	case ast.Var:
		b, next := u.apply(a)
		if a != b || u != next {
			return next.plugNamespaced(b, caller)
		}
		return u.namespaceVar(b, caller)
	case *ast.Array:
		if a.IsGround() {
			return a
		}
		cpy := *a
		arr := make([]*ast.Term, v.Len())
		for i := range arr {
			arr[i] = u.plugNamespaced(v.Elem(i), caller)
		}
		cpy.Value = ast.NewArray(arr...)
		return &cpy
	case ast.Object:
		if a.IsGround() {
			return a
		}
		cpy := *a
		cpy.Value, _ = v.Map(func(k, v *ast.Term) (*ast.Term, *ast.Term, error) {
			return u.plugNamespaced(k, caller), u.plugNamespaced(v, caller), nil
		})
		return &cpy
	case ast.Set:
		if a.IsGround() {
			return a
		}
		cpy := *a
		cpy.Value, _ = v.Map(func(x *ast.Term) (*ast.Term, error) {
			return u.plugNamespaced(x, caller), nil
		})
		return &cpy
	case ast.Ref:
		cpy := *a
		ref := make(ast.Ref, len(v))
		for i := range ref {
			ref[i] = u.plugNamespaced(v[i], caller)
		}
		cpy.Value = ref
		return &cpy
	}
	return a
}

func (u *bindings) bind(a *ast.Term, b *ast.Term, other *bindings, und *undo) {
	k := unique.Make(a.Value.(ast.Var))
	v := bindingKeyValue{key: a, val: b, bindings: other}

	if u.values == nil {
		u.values = make(map[unique.Handle[ast.Var]]bindingKeyValue, 6)
	}
	u.values[k] = v

	und.k = a
	und.u = u
}

func (u *bindings) apply(a *ast.Term) (*ast.Term, *bindings) {
	if u != nil && a != nil {
		// Early exit for non-var terms. Only vars are bound in the binding list,
		// so the lookup below will always fail for non-var terms. In some cases,
		// the lookup may be expensive as it has to hash the term (which for large
		// inputs can be costly).
		if v, ok := a.Value.(ast.Var); ok {
			if val, ok := u.values[unique.Make(v)]; ok {
				return val.bindings.apply(val.val)
			}
		}
	}
	return a, u
}

func (u *bindings) String() string {
	if u == nil {
		return "()"
	}
	var buf []string
	for _, v := range u.values {
		buf = append(buf, fmt.Sprintf("%v: %v", v.key, v.bindings))
	}

	return fmt.Sprintf("({%v}, %v)", strings.Join(buf, ", "), u.id)
}

func (u *bindings) namespaceVar(v *ast.Term, caller *bindings) *ast.Term {
	name, ok := v.Value.(ast.Var)
	if !ok {
		panic("illegal value")
	}
	if caller != nil && caller != u {
		// Root documents (i.e., data, input) should never be namespaced because they
		// are globally unique.
		if !ast.RootDocumentNames.Contains(v) {
			return ast.VarTerm(string(name) + strconv.FormatUint(u.id, 10))
		}
	}
	return v
}

type namespacingVisitor struct {
	b      *bindings
	caller *bindings
}

func (vis namespacingVisitor) Visit(x any) bool {
	switch x := x.(type) {
	case *ast.ArrayComprehension:
		x.Term = vis.namespaceTerm(x.Term)
		vis := ast.NewGenericVisitor(vis.Visit)
		for _, expr := range x.Body {
			vis.Walk(expr)
		}
		return true
	case *ast.SetComprehension:
		x.Term = vis.namespaceTerm(x.Term)
		vis := ast.NewGenericVisitor(vis.Visit)
		for _, expr := range x.Body {
			vis.Walk(expr)
		}
		return true
	case *ast.ObjectComprehension:
		x.Key = vis.namespaceTerm(x.Key)
		x.Value = vis.namespaceTerm(x.Value)
		vis := ast.NewGenericVisitor(vis.Visit)
		for _, expr := range x.Body {
			vis.Walk(expr)
		}
		return true
	case *ast.Expr:
		switch terms := x.Terms.(type) {
		case []*ast.Term:
			for i := 1; i < len(terms); i++ {
				terms[i] = vis.namespaceTerm(terms[i])
			}
		case *ast.Term:
			x.Terms = vis.namespaceTerm(terms)
		}
		for _, w := range x.With {
			w.Target = vis.namespaceTerm(w.Target)
			w.Value = vis.namespaceTerm(w.Value)
		}
	}
	return false
}

func (vis namespacingVisitor) namespaceTerm(a *ast.Term) *ast.Term {
	switch v := a.Value.(type) {
	case ast.Var:
		return vis.b.namespaceVar(a, vis.caller)
	case *ast.Array:
		if a.IsGround() {
			return a
		}
		cpy := *a
		arr := make([]*ast.Term, v.Len())
		for i := range arr {
			arr[i] = vis.namespaceTerm(v.Elem(i))
		}
		cpy.Value = ast.NewArray(arr...)
		return &cpy
	case ast.Object:
		if a.IsGround() {
			return a
		}
		cpy := *a
		cpy.Value, _ = v.Map(func(k, v *ast.Term) (*ast.Term, *ast.Term, error) {
			return vis.namespaceTerm(k), vis.namespaceTerm(v), nil
		})
		return &cpy
	case ast.Set:
		if a.IsGround() {
			return a
		}
		cpy := *a
		cpy.Value, _ = v.Map(func(x *ast.Term) (*ast.Term, error) {
			return vis.namespaceTerm(x), nil
		})
		return &cpy
	case ast.Ref:
		cpy := *a
		ref := make(ast.Ref, len(v))
		for i := range ref {
			ref[i] = vis.namespaceTerm(v[i])
		}
		cpy.Value = ref
		return &cpy
	}
	return a
}
