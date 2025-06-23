package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EvaluateCondition is a helper for applyFilter
func EvaluateCondition(rowValue interface{}, operator string, conditionValue interface{}) (bool, error) {
	// This is a complex part. Needs robust type handling and operator logic.
	// TODO: Implement proper type conversion and comparison for various operators and types.
	// Example for basic equality on strings and numbers:
	rowStr := fmt.Sprintf("%v", rowValue)
	condStr := fmt.Sprintf("%v", conditionValue)

	switch operator {
	case "=":
		// Attempt numeric comparison if possible, otherwise string
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat == condFloat, nil
		}
		return rowStr == condStr, nil
	case "!=":
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat != condFloat, nil
		}
		return rowStr != condStr, nil
	case ">":
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat > condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "<":
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat < condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case ">=":
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat >= condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "<=":
		rowFloat, rOk := ConvertToFloat64(rowValue)
		condFloat, cOk := ConvertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat <= condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "CONTAINS":
		return strings.Contains(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "NOT_CONTAINS":
		return !strings.Contains(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "STARTS_WITH":
		return strings.HasPrefix(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "ENDS_WITH":
		return strings.HasSuffix(strings.ToLower(rowStr), strings.ToLower(condStr)), nil

		// TODO: Add support for IN, NOT IN, IS NULL, IS NOT NULL, etc.
	default:
		return false, fmt.Errorf("unsupported filter operator: %s", operator)
	}
}

// ConvertToFloat64 Helper to convert interface{} to float64 for comparisons
func ConvertToFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
