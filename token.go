package main

// Token represents a lexical token.
type Token int

const (
    // Special tokens
    ILLEGAL Token = iota
    EOF
    WS

    // Literals
    FIELD 
    NUMBER

    // Operators
    ADD         // +
    SUBTRACT    // -
    MULTIPLY    // *
    DIVIDE      // /
    EQ          // =
    GT          // >
    LT          // <
    GTE         // >=
    LTE         // <=
    AND         // and

    // Misc characters
    COMMA    // ,

    // Keywords
    MAP
    REDUCE
    ON
    WHERE
    IN
)
