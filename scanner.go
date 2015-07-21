package main

import (
    "bufio"
    "bytes"
    "io"
    "regexp"
    "strings"
)

// Scanner represents a lexical scanner.
type Scanner struct {
    r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
    return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
    // Read the next rune.
    ch := s.read()

    // If we see whitespace then consume all contiguous whitespace.
    // If we see a letter then consume as an ident or reserved word.
    // If we see a digit then consume as a number.
    if isWhitespace(ch) {
        s.unread()
        return s.scanWhitespace()
    } else if (ch == eof) {
        return EOF, ""
    } else if (ch == ',') {
        return COMMA, string(ch)
    } else if (isValidCh(ch)) {
        s.unread()
        return s.scanToken()
    }

    return ILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
    // Create a buffer and read the current character into it.
    var buf bytes.Buffer
    buf.WriteRune(s.read())

    // Read every subsequent whitespace character into the buffer.
    // Non-whitespace characters and EOF will cause the loop to exit.
    for {
        if ch := s.read(); ch == eof {
            break
        } else if !isWhitespace(ch) {
            s.unread()
            break
        } else {
            buf.WriteRune(ch)
        }
    }

    return WS, buf.String()
}

// scanToken consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanToken() (tok Token, lit string) {
    // Create a buffer and read the current character into it.
    var buf bytes.Buffer
    buf.WriteRune(s.read())

    // Read every subsequent ident character into the buffer.
    // Non-ident characters and EOF will cause the loop to exit.
    for {
        if ch := s.read(); ch == eof {
            break
        } else if isWhitespace(ch) || !isValidCh(ch) {
            s.unread()
            break
        } else {
            _, _ = buf.WriteRune(ch)
        }
    }

    // If the string matches a keyword then return that keyword.
    switch strings.ToUpper(buf.String()) {
    case "MAP":
        return MAP, buf.String()
    case "REDUCE":
        return REDUCE, buf.String()
    case "ON":
        return ON, buf.String()
    case "WHERE":
        return WHERE, buf.String()
    case "AND":
        return AND, buf.String()
    }

    // Match operators
    switch buf.String() {
    case "<":
        return LT, buf.String()
    case "<=":
        return LTE, buf.String()
    case ">":
        return GT, buf.String()
    case ">=":
        return GTE, buf.String()
    case "=":
        return EQ, buf.String()
    }

    // Check for string literal
    var field string = buf.String()
    if field[0] == '"' && field[len(field)-1] == '"' {
        field = field[1:len(field)-1]
        return FIELD, field
    }

    // Check for number
    if match, _ := regexp.MatchString("([0-9.]+)", buf.String()); match {
        return NUMBER, buf.String()
    }

    return ILLEGAL, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
    ch, _, err := s.r.ReadRune()
    if err != nil {
        return eof
    }
    return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

func isOperator(ch rune) bool { return (ch == '<' || ch == '>' || ch == '=' || ch == '+' || ch == '-' || ch == '/' || ch == '*') }

func isValidCh(ch rune) bool { 
    return isWhitespace(ch) || isLetter(ch) || isDigit(ch) || isOperator(ch) || ch == '_' || ch == '"'
}

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
