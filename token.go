package main

// Token represents a lexical token.
type Token int

const (
    // Special tokens
    ILLEGAL Token = iota
    EOF
    WS

    // Literals
    STRING
    NUMBER
    IDENT
    
    // Types
    TYPE_INT
    TYPE_NIL
    TYPE_STRING
    TYPE_FLOAT
    TYPE_PROPERTY
    TYPE_LIST

    // Operators
    ADD         // +
    SUBTRACT    // -
    MULTIPLY    // *
    DIVIDE      // /
    EQ          // =
    NOT_EQ      // !=
    GT          // >
    LT          // <
    GTE         // >=
    LTE         // <=
    AND         // and

    // Aggregate methods
    SUM

    // Misc characters
    COMMA    // ,

    // Keywords
    MAP
    REDUCE
    ON
    WHERE
    IN
)

