// Copyright 2025 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
)

func builtinTemplateString(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	arr, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	w := opWriterPool.Get()
	w.singleOutput = true

	w.Prepare(nil, arr, bctx.Location, outputFunc)
	defer func() {
		w.Clear()
		opWriterPool.Put(w)
	}()

	if err := builtinPrintCrossProductOperands(w, 0); err != nil {
		return err
	}

	return iter(ast.InternedTermBytes(w.Bytes()))
}

func outputFunc(w *operandWriter) error {
	w.count += 1
	// Precautionary run-time assertion that template-strings can't produce multiple outputs;
	// e.g. for custom relation type built-ins not known at compile-time.
	if w.count > 1 {
		return Halt{Err: &Error{
			Code:     ConflictErr,
			Location: w.loc,
			Message:  "template-strings must not produce multiple outputs",
		}}
	}
	return nil
}

func init() {
	RegisterBuiltinFunc(ast.InternalTemplateString.Name, builtinTemplateString)
}
