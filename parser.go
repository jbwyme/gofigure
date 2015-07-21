package main

import (
    "fmt"
    "io"
    "strconv"
)

type Condition struct {
    left    string
    op      Token
    right   float64
}

type MapStatement struct {
    Fields      []string
    Conditions  []Condition  
}

type ReduceStatement struct {
    Key     string
    Conditions  []Condition
}    

// Parser represents a parser.
type Parser struct {
    s   *Scanner
    buf struct {
        tok Token  // last read token
        lit string // last read literal
        n   int    // buffer size (max=1)
    }
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
    return &Parser{s: NewScanner(r)}
}

// Parse parses a MAP REDUCE statement.
func (p *Parser) Parse() (*MapStatement, *ReduceStatement, error) {
    ms := &MapStatement{}
    rs := &ReduceStatement{}

    // First token should be a "MAP" keyword.
    if tok, lit := p.scanIgnoreWhitespace(); tok != MAP {
        return nil, nil, fmt.Errorf("found %q, expected MAP", lit)
    }

    // Next we should loop over all our comma-delimited fields.
    for {
        // Read a field.
        tok, lit := p.scanIgnoreWhitespace()
        if tok != FIELD {
            return nil, nil, fmt.Errorf("found %q, expected field", lit)
        }
        ms.Fields = append(ms.Fields, lit)

        // If the next token is not a comma then break the loop.
        if tok, _ := p.scanIgnoreWhitespace(); tok != COMMA {
            p.unscan()
            break
        }
    }
    
    // Check for conditionals in MAP
    if tok, _ := p.scan(); tok == WHERE {
        // Next we should loop over conditions
        for {
            condition := Condition{}

            // Read a left side of condition
            tok, lit := p.scanIgnoreWhitespace()
            if tok != FIELD {
                return nil, nil, fmt.Errorf("found %q, expected field", lit)
            }
            condition.left = lit

            // Read operator
            tok, lit = p.scanIgnoreWhitespace()
            if !(tok == GT || tok == EQ) {
                return nil, nil, fmt.Errorf("found %q, expected operator", lit)
            }
            condition.op = tok
    
            // Read operand
            tok, lit = p.scanIgnoreWhitespace()
            if tok != NUMBER {
                return nil, nil, fmt.Errorf("found %q, expected operand", lit)
            }
            if i, err := strconv.ParseFloat(lit, 64); err != nil {
                return nil, nil, fmt.Errorf("operand must be type Float", lit)
            } else {
                condition.right = i
            }
            
            ms.Conditions = append(ms.Conditions, condition)
   
            if tok, _ := p.scanIgnoreWhitespace(); tok != AND {
                p.unscan()
                break
            }
        }
    } else {
        p.unscan()
    }

    // Next we should see the "REDUCE" keyword.
    if tok, lit := p.scanIgnoreWhitespace(); tok != REDUCE {
        return nil, nil, fmt.Errorf("found %q, expected REDUCE", lit)
    }
    
    // Next we should see the "ON" keyword.
    if tok, lit := p.scanIgnoreWhitespace(); tok != ON {
        return nil, nil, fmt.Errorf("found %q, expected ON", lit)
    }

    // Finally we should read the reduce key.
    tok, lit := p.scanIgnoreWhitespace()
    if tok != FIELD {
        return nil, nil, fmt.Errorf("found %q, expected reduce key", lit)
    }
    rs.Key = lit

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
