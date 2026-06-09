package ast

import (
	"encoding"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"

	astJSON "github.com/open-policy-agent/opa/v1/ast/json"
	"github.com/open-policy-agent/opa/v1/util"
)

var (
	_ json.MarshalerTo = &Term{}
	_ json.Unmarshaler = &LogicalOr{}
	_ json.MarshalerTo = &LogicalOr{}
	_ json.MarshalerTo = &Not{}
	_ json.MarshalerTo = &Array{}
	_ json.MarshalerTo = &set{}
	_ json.MarshalerTo = &object{}
	_ json.MarshalerTo = &TemplateString{}
	_ json.MarshalerTo = &Ref{}
)

func WriteTokens(e *jsontext.Encoder, tokens ...jsontext.Token) error {
	for _, t := range tokens {
		if err := e.WriteToken(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Term) MarshalJSONTo(e *jsontext.Encoder) (err error) {
	e.WriteToken(jsontext.BeginObject)

	includeLocation := astJSON.GetOptions().MarshalOptions.IncludeLocation
	if t.Location != nil && includeLocation.Term {
		e.WriteToken(jsontext.String("location"))
		t.Location.MarshalJSONTo(e)
	}

	e.WriteToken(jsontext.String("type"))
	e.WriteToken(jsontext.String(ValueName(t.Value)))

	e.WriteToken(jsontext.String("value"))
	if err = marshalValueTo(e, t.Value); err != nil {
		return fmt.Errorf("failed to marshal term of %s: %w", ValueName(t.Value), err)
	}

	return e.WriteToken(jsontext.EndObject)
}

func marshalValueTo(e *jsontext.Encoder, val Value) (err error) {
	switch v := val.(type) {
	case Var:
		err = e.WriteToken(jsontext.String(string(v)))
	case Null:
		err = e.WriteValue([]byte("{}"))
	case json.MarshalerTo:
		err = v.MarshalJSONTo(e)
	case encoding.TextAppender:
		buf, _ := v.AppendText(e.AvailableBuffer())
		err = e.WriteValue(buf)
	default:
		err = json.MarshalEncode(e, v)
	}
	return err
}

// MarshalJSON returns the JSON encoding of the term.
func (term *Term) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(term)
}

func (r Ref) MarshalJSONTo(e *jsontext.Encoder) (err error) {
	return util.WriteMarshalerToArray(e, r)
}

func (t *TemplateString) MarshalJSONTo(e *jsontext.Encoder) (err error) {
	e.WriteToken(jsontext.BeginObject)
	e.WriteToken(jsontext.String("parts"))
	e.WriteToken(jsontext.BeginArray)
	for _, p := range t.Parts {
		switch v := p.(type) {
		case *Expr:
			v.MarshalJSONTo(e)
		case *Term:
			v.MarshalJSONTo(e)
		}
	}
	e.WriteToken(jsontext.EndArray)

	e.WriteToken(jsontext.String("multi_line"))
	e.WriteToken(jsontext.Bool(t.MultiLine))

	return e.WriteToken(jsontext.EndObject)
}

func (n *Not) MarshalJSONTo(e *jsontext.Encoder) error {
	e.WriteToken(jsontext.BeginObject)
	e.WriteToken(jsontext.String("type"))
	e.WriteToken(jsontext.String("not"))

	e.WriteToken(jsontext.String("body"))
	if err := n.Body.MarshalJSONTo(e); err != nil {
		return err
	}

	e.WriteToken(jsontext.String("explicit_body"))
	e.WriteToken(jsontext.Bool(n.ExplicitBody))

	if astJSON.GetOptions().MarshalOptions.IncludeLocation.Not && n.Location != nil {
		e.WriteToken(jsontext.String("location"))
		n.Location.MarshalJSONTo(e)
	}

	return e.WriteToken(jsontext.EndObject)
}

func (n *Not) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(n)
}

func (obj *object) MarshalJSONTo(e *jsontext.Encoder) error {
	e.WriteToken(jsontext.BeginArray)
	for _, node := range obj.sortedKeys() {
		e.WriteToken(jsontext.BeginArray)
		node.key.MarshalJSONTo(e)
		node.value.MarshalJSONTo(e)
		e.WriteToken(jsontext.EndArray)
	}
	return e.WriteToken(jsontext.EndArray)
}

func (l *lazyObj) MarshalJSON() ([]byte, error) {
	return l.force().(*object).MarshalJSON()
}

// MarshalJSON returns JSON encoded bytes representing obj.
func (obj *object) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(obj)
}

func (a *Array) MarshalJSONTo(e *jsontext.Encoder) error {
	return util.WriteMarshalerToArray(e, a.elems)
}

// MarshalJSON returns JSON encoded bytes representing arr.
func (arr *Array) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(arr)
}

func (s *set) MarshalJSONTo(e *jsontext.Encoder) error {
	return util.WriteMarshalerToArray(e, s.sortedKeys())
}

// MarshalJSON returns JSON encoded bytes representing s.
func (s *set) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(s)
}

func (o *LogicalOr) MarshalJSON() ([]byte, error) {
	return util.MarshalMarshalerTo(o)
}

func (o *LogicalOr) MarshalJSONTo(e *jsontext.Encoder) error {
	e.WriteToken(jsontext.BeginObject)

	e.WriteToken(jsontext.String("type"))
	e.WriteToken(jsontext.String("or"))

	e.WriteToken(jsontext.String("lhs"))
	if err := o.Lhs.MarshalJSONTo(e); err != nil {
		return err
	}

	e.WriteToken(jsontext.String("rhs"))
	if err := o.Rhs.MarshalJSONTo(e); err != nil {
		return err
	}

	if o.ExplicitLhs {
		e.WriteToken(jsontext.String("explicit_lhs"))
		e.WriteToken(jsontext.True)
	}
	if o.ExplicitRhs {
		e.WriteToken(jsontext.String("explicit_rhs"))
		e.WriteToken(jsontext.True)
	}

	if astJSON.GetOptions().MarshalOptions.IncludeLocation.Or && o.Location != nil {
		e.WriteToken(jsontext.String("location"))
		if err := o.Location.MarshalJSONTo(e); err != nil {
			return err
		}
	}

	return e.WriteToken(jsontext.EndObject)
}

func (o *LogicalOr) UnmarshalJSON(bs []byte) error {
	v := map[string]any{}
	if err := util.UnmarshalJSON(bs, &v); err != nil {
		return err
	}
	return unmarshalLogical("or", &o.Lhs, &o.Rhs, &o.ExplicitLhs, &o.ExplicitRhs, v)
}

// UnmarshalJSON parses the byte array and stores the result in term.
// Specialized unmarshalling is required to handle Value and Location.
func (term *Term) UnmarshalJSON(bs []byte) error {
	v := map[string]any{}
	if err := util.UnmarshalJSON(bs, &v); err != nil {
		return err
	}
	val, err := unmarshalValue(v)
	if err != nil {
		return err
	}
	term.Value = val

	if loc, ok := v["location"].(map[string]any); ok {
		term.Location = &Location{}
		err := unmarshalLocation(term.Location, loc)
		if err != nil {
			return err
		}
	}
	return nil
}
