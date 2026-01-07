// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"io"
	"slices"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
	"github.com/open-policy-agent/opa/v1/topdown/print"
	"github.com/open-policy-agent/opa/v1/util"
)

var (
	opWriterPool = util.NewSyncPool[operandWriter]()
	newLineBytes = []byte{'\n'}
)

func NewPrintHook(w io.Writer) print.Hook {
	return printHook{w: w}
}

type printHook struct {
	w io.Writer
}

type operandWriter struct {
	bufs         [][]byte
	delim        []byte
	operands     *ast.Array
	loc          *ast.Location
	f            func(*operandWriter) error
	count        byte
	singleOutput bool
}

func (ow *operandWriter) Prepare(
	delim []byte,
	operands *ast.Array,
	loc *ast.Location,
	f func(*operandWriter) error,
) *operandWriter {
	lo := operands.Len()
	if ow.bufs == nil {
		ow.bufs = make([][]byte, lo)
	} else if lo > len(ow.bufs) {
		ow.bufs = append(ow.bufs, make([][]byte, lo-len(ow.bufs))...)
	}

	l := max(0, len(delim)*(lo-1))
	for i := range lo {
		l += operands.Elem(i).StringLength()
		if ow.bufs[i] == nil {
			ow.bufs[i] = make([]byte, 0, l)
			continue
		}

		ow.bufs[i] = slices.Grow(ow.bufs[i], l)[:0]
	}

	ow.delim = delim
	ow.operands = operands
	ow.loc = loc
	ow.f = f
	ow.count = 0

	return ow
}

func (ow *operandWriter) Clear() *operandWriter {
	if len(ow.bufs) > 0 {
		ow.bufs[0] = ow.bufs[0][:0]
	}

	return ow
}

func (ow *operandWriter) String() string {
	return string(ow.Bytes())
}

func (ow *operandWriter) Bytes() []byte {
	numOps := ow.operands.Len()
	if numOps == 0 {
		return nil
	}
	if numOps == 1 {
		if len(ow.bufs[0]) == 0 {
			return nil
		}
		return ow.bufs[0]
	}

	lenRest := len(ow.delim) * max(numOps-1, 0)
	for i := range numOps {
		lenRest += len(ow.bufs[i])
	}

	if lenRest == 0 {
		return ow.bufs[0]
	}

	if ow.singleOutput {
		// If we don't allow multiple outputs, we can optimize some by
		// reusing the the first buf to append the rest of the operand text into..
		ow.bufs[0] = slices.Grow(ow.bufs[0], lenRest)
		if len(ow.delim) == 0 {
			for i := 1; i < numOps; i++ {
				ow.bufs[0] = append(ow.bufs[0], ow.bufs[i]...)
			}
			return ow.bufs[0]
		}

		for i := 1; i < numOps; i++ {
			ow.bufs[0] = append(ow.bufs[0], ow.delim...)
			ow.bufs[0] = append(ow.bufs[0], ow.bufs[i]...)
		}

		return ow.bufs[0]
	}

	// If *do* allow multiple outputs (print apparently does), we can't overwrite
	// the first buf until we've constructed the full output. So we need to allocate
	// a new buf here.
	totLen := lenRest + len(ow.bufs[0])
	buf := append(make([]byte, 0, totLen), ow.bufs[0]...)
	for i := 1; i < numOps; i++ {
		buf = append(buf, ow.delim...)
		buf = append(buf, ow.bufs[i]...)
	}

	return buf
}

func (h printHook) Print(_ print.Context, msg string) error {
	if len(msg) > 0 {
		if _, err := h.w.Write(util.StringToByteSlice(msg)); err != nil {
			return err
		}
	}
	_, err := h.w.Write(newLineBytes)
	return err
}

var delim = []byte(" ")

func builtinPrint(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	if bctx.PrintHook == nil {
		return iter(nil)
	}

	arr, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Avoid bctx escaping to heap by being referenced in the closure below.
	// These variables are interface/pointer types and do not escape.
	ctx, hook := bctx.Context, bctx.PrintHook

	w := opWriterPool.Get().Prepare(delim, arr, bctx.Location, func(w *operandWriter) error {
		pctx := print.Context{Context: ctx, Location: w.loc}
		return hook.Print(pctx, w.String())
	})
	defer func() {
		opWriterPool.Put(w)
	}()

	if err := builtinPrintCrossProductOperands(w, 0); err != nil {
		return err
	}

	return iter(nil)
}

func builtinPrintCrossProductOperands(w *operandWriter, i int) error {
	numOperands := w.operands.Len()
	if numOperands == 0 || i >= numOperands {
		return w.f(w)
	}

	operand := w.operands.Elem(i)
	// We allow primitives ...
	switch x := operand.Value.(type) {
	case ast.String:
		w.bufs[i] = append(w.bufs[i][:0], string(x)...)
		return builtinPrintCrossProductOperands(w, i+1)
	case ast.Number, ast.Boolean, ast.Null:
		w.bufs[i], _ = operand.AppendText(w.bufs[i])
		return builtinPrintCrossProductOperands(w, i+1)
	}

	// ... but all other operand types must be sets.
	xs, ok := operand.Value.(ast.Set)
	if !ok {
		return Halt{Err: internalErr(w.loc, "illegal argument type: "+ast.ValueName(operand.Value))}
	}

	if xs.Len() == 0 {
		w.bufs[i] = append(w.bufs[i][:0], "<undefined>"...)
		return builtinPrintCrossProductOperands(w, i+1)
	}

	return xs.Iter(func(x *ast.Term) error {
		switch v := x.Value.(type) {
		case ast.String:
			w.bufs[i] = append(w.bufs[i][:0], v...)
		default:
			var err error
			if w.bufs[i], err = x.AppendText(w.bufs[i][:0]); err != nil {
				return err
			}
		}
		return builtinPrintCrossProductOperands(w, i+1)
	})
}

func init() {
	RegisterBuiltinFunc(ast.InternalPrint.Name, builtinPrint)
}
