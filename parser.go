package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

type ItrOperator string

const (
	OP_IN = "IN"
)

type AggregateMethod string

const (
	AGG_SUM = "SUM"
)

type IField interface {
	GetVal() interface{}
	SetVal(v interface{})
	GetType() Token
	SetType(t Token)
	GetName() string
}

type Field struct {
	Name      string
	Type      Token
	Val       interface{}
	IntVal    int
	FloatVal  float64
	StringVal string
	ListVal   []interface{}
}

func (f *Field) GetVal() interface{} {
	return f.Val
}

func (f *Field) SetVal(v interface{}) {
	f.Val = v
}

func (f *Field) GetType() Token {
	return f.Type
}

func (f *Field) SetType(t Token) {
	f.Type = t
}

func (f *Field) GetName() string {
	return f.Name
}

type FieldItr struct {
	Field
	Operator   ItrOperator
	Collection IField
}

type Condition struct {
	left  IField
	op    Token
	right IField
}

type Aggregator struct {
	Field
	Target IField
	Method AggregateMethod
}

type IStatement interface {
	GetFields() []IField
	AddField(f IField)
	GetConditions() []Condition
	AddCondition(c Condition)
}

type Statement struct {
	Fields     []IField
	Conditions []Condition
}

func (s *Statement) GetFields() []IField {
	return s.Fields
}

func (s *Statement) AddField(f IField) {
	if s.Fields == nil {
		s.Fields = make([]IField, 0)
	}
	s.Fields = append(s.Fields, f)
}

func (s *Statement) GetConditions() []Condition {
	return s.Conditions
}

func (s *Statement) AddCondition(c Condition) {
	s.Conditions = append(s.Conditions, c)
}

type ReduceStatement struct {
	Statement
	Key string
}

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func createField(fieldType Token, name string) *Field {
	n := &Field{Type: fieldType, Name: name}
	switch n.Type {
	case TYPE_STRING, TYPE_PROPERTY:
		n.StringVal = name
	case TYPE_INT:
		if v, err := strconv.Atoi(name); err != nil {
			panic(err)
		} else {
			n.IntVal = v
		}
	case TYPE_FLOAT:
		if v, err := strconv.ParseFloat(name, 64); err != nil {
			panic(err)
		} else {
			n.FloatVal = v
		}
	}
	return n
}

func tokenToField(tok Token, lit string) (*Field, error) {
	switch tok {
	case NUMBER:
		return createField(TYPE_FLOAT, lit), nil
	case STRING:
		return createField(TYPE_STRING, lit), nil
	case IDENT:
		return createField(TYPE_PROPERTY, lit), nil
	default:
		return &Field{}, errors.New(fmt.Sprintf("Unable to determine type of Field to created for '%s'", lit))
	}
}

func (p *Parser) parseField(stmt IStatement) (IField, error) {
	tok, field := p.scanIgnoreWhitespace()
	if tok == SUM {
		if targetField, err := p.parseField(stmt); err == nil {
			return &Aggregator{Field: Field{Name: targetField.GetName()}, Method: AGG_SUM, Target: targetField}, nil
		} else {
			return nil, err
		}
	} else if tok == IDENT {
		fieldNode := createField(TYPE_PROPERTY, field)
		tok, _ = p.scanIgnoreWhitespace()
		if tok == IN {
			if collectionField, err := p.parseField(stmt); err == nil {
				return &FieldItr{Field: *fieldNode, Collection: collectionField, Operator: OP_IN}, nil
			} else {
				return nil, err
			}
		} else if tok == COMMA {
			return fieldNode, nil
		} else {
			p.unscan()
			return fieldNode, nil
		}
	} else {
		return nil, fmt.Errorf("Found %d, expected IDENT or SUM", field)
	}
}

func (p *Parser) parseFields(stmt IStatement) error {
	for true {
		var tok Token
		tok, _ = p.scanIgnoreWhitespace()
		if tok == EOF || tok == REDUCE || tok == ON || tok == WHERE {
			p.unscan()
			break
		} else {
			p.unscan()
			if f, err := p.parseField(stmt); err == nil {
				stmt.AddField(f)
			} else {
				panic(err)
			}
		}
	}
	return nil
}

func (p *Parser) parseWhere(stmt IStatement) error {
	for {
		condition := Condition{}

		// Read a left side of condition
		tok, lit := p.scanIgnoreWhitespace()
		l, err := tokenToField(tok, lit)
		if err != nil {
			return err
		} else {
			condition.left = l
		}

		// Read operator
		tok, lit = p.scanIgnoreWhitespace()
		if !(tok == GT || tok == GTE || tok == EQ || tok == NOT_EQ || tok == LT || tok == LTE) {
			return fmt.Errorf("found %q, expected operator", lit)
		}
		condition.op = tok

		// Read operand
		tok, lit = p.scanIgnoreWhitespace()
		r, err := tokenToField(tok, lit)
		if err != nil {
			return err
		} else {
			condition.right = r
		}

		stmt.AddCondition(condition)

		if tok, _ := p.scanIgnoreWhitespace(); tok != AND {
			p.unscan()
			break
		}
	}
	return nil
}

// Parse parses a MAP REDUCE statement.
func (p *Parser) Parse() (*Statement, *ReduceStatement, error) {
	ms := &Statement{}
	rs := &ReduceStatement{}

	// First token should be a "MAP" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != MAP {
		return nil, nil, fmt.Errorf("found %q, expected MAP", lit)
	}

	// Next we should loop over all our comma-delimited fields for the MAP statement
	if err := p.parseFields(ms); err != nil {
		return nil, nil, err
	}

	// Check for conditionals in MAP
	if tok, _ := p.scan(); tok == WHERE {
		p.parseWhere(ms)
	} else {
		p.unscan()
	}

	// Next we should see the "REDUCE" keyword.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != EOF {
		if tok != REDUCE {
			return nil, nil, fmt.Errorf("found %s, expected REDUCE", lit)
		}

		if err := p.parseFields(rs); err != nil {
			return nil, nil, err
		}

		// Check for conditionals in REDUCE
		if tok, _ := p.scan(); tok == WHERE {
			p.parseWhere(rs)
		} else {
			p.unscan()
		}

		if tok, lit := p.scanIgnoreWhitespace(); tok != ON {
			return nil, nil, fmt.Errorf("found %s, expected ON", lit)
		}

		// Finally we should read the reduce key.
		tok, lit := p.scanIgnoreWhitespace()
		if tok != IDENT {
			return nil, nil, fmt.Errorf("found %s, expected reduce key", lit)
		}
		rs.Key = lit
	}

	// Return the successfully parsed statement.
	return ms, rs, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
