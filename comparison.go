package main

import "strconv"

func compareIntToInt(left int, right int, op Token) bool {
    switch(op) {
    case EQ:
        return left == right
    case NOT_EQ:
        return left != right
    case GT: 
        return left > right
    case GTE:
        return left >= right
    case LT:
        return left < right
    case LTE:
        return left <= right
    default:
        return false
    }
}

func compareFloatToFloat(left float64, right float64, op Token) bool {
    switch(op) {
    case EQ:
        return left == right
    case NOT_EQ:
        return left != right
    case GT: 
        return left > right
    case GTE:
        return left >= right
    case LT:
        return left < right
    case LTE:
        return left <= right
    default:
        return false
    }
}

func compareStringToString(left string, right string, op Token) bool {
    switch(op) {
    case EQ:
        return left == right
    case NOT_EQ:
        return left != right
    case GT: 
        return left > right
    case GTE:
        return left >= right
    case LT:
        return left < right
    case LTE:
        return left <= right
    default:
        return false
    }
}

func compareIntToFloat(left int, right float64, op Token) bool {
    return compareFloatToFloat(float64(left), right, op)
}

func compareFloatToInt(left float64, right int, op Token) bool {
    return compareFloatToFloat(left, float64(right), op)
}

func compareIntToString(left int, right string, op Token) bool {
    if rightInt, err := strconv.Atoi(right); err == nil {
        return compareIntToInt(left, rightInt, op)
    } else {
        return false
    }
}

func compareStringToInt(left string, right int, op Token) bool {
    if leftInt, err := strconv.Atoi(left); err == nil {
        return compareIntToInt(leftInt, right, op)
    } else {
        return false
    }
}

func compareFloatToString(left float64, right string, op Token) bool {
    if rightFloat, err := strconv.ParseFloat(right, 64); err == nil {
        return compareFloatToFloat(left, rightFloat, op)
    } else {
        return false
    }
}

func compareStringToFloat(left string, right float64, op Token) bool {
    if leftFloat, err := strconv.ParseFloat(left, 64); err == nil {
        return compareFloatToFloat(leftFloat, right, op)
    } else {
        return false
    }
} 


