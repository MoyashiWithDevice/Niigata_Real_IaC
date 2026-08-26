package condition

import (
	"fmt"
	"regexp"
	"strings"
)

// ToFloat64 converts a value to float64 for comparison.
func ToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

// Evaluate evaluates a condition against a value using the given operator.
// Supported operators: eq, ne, in, nin, gt, ge, lt, le, contains,
// starts_with, ends_with, matches, defined, undefined.
func Evaluate(value, expected interface{}, operator string) (bool, error) {
	switch operator {
	case "eq":
		return evaluateEq(value, expected)
	case "ne":
		result, err := evaluateEq(value, expected)
		return !result, err
	case "in":
		return evaluateIn(value, expected)
	case "nin":
		result, err := evaluateIn(value, expected)
		return !result, err
	case "gt":
		return evaluateNumeric(value, expected, func(v, e float64) bool { return v > e })
	case "ge":
		return evaluateNumeric(value, expected, func(v, e float64) bool { return v >= e })
	case "lt":
		return evaluateNumeric(value, expected, func(v, e float64) bool { return v < e })
	case "le":
		return evaluateNumeric(value, expected, func(v, e float64) bool { return v <= e })
	case "contains":
		return evaluateContains(value, expected)
	case "starts_with":
		return evaluateStartsWith(value, expected)
	case "ends_with":
		return evaluateEndsWith(value, expected)
	case "matches":
		return evaluateMatches(value, expected)
	case "defined":
		return evaluateDefined(value, expected)
	case "undefined":
		result, err := evaluateDefined(value, expected)
		return !result, err
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// evaluateEq evaluates equality.
func evaluateEq(value, expected interface{}) (bool, error) {
	if value == nil && expected == nil {
		return true, nil
	}
	if value == nil || expected == nil {
		return false, nil
	}
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", expected), nil
}

// evaluateIn evaluates if value is in a list.
func evaluateIn(value interface{}, list interface{}) (bool, error) {
	if list == nil {
		return false, nil
	}

	listSlice, ok := list.([]interface{})
	if !ok {
		return false, fmt.Errorf("in operator requires a list, got %T", list)
	}

	for _, item := range listSlice {
		matches, err := evaluateEq(value, item)
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}

	return false, nil
}

// evaluateNumeric evaluates a numeric comparison using the given comparator.
func evaluateNumeric(value, expected interface{}, compare func(v, e float64) bool) (bool, error) {
	v, ok := ToFloat64(value)
	if !ok {
		return false, fmt.Errorf("cannot compare non-numeric value: %v", value)
	}
	ev, ok := ToFloat64(expected)
	if !ok {
		return false, fmt.Errorf("cannot compare with non-numeric value: %v", expected)
	}
	return compare(v, ev), nil
}

// evaluateContains evaluates string contains or slice contains.
func evaluateContains(value, expected interface{}) (bool, error) {
	// Handle string contains
	if str, ok := value.(string); ok {
		if expectedStr, ok := expected.(string); ok {
			return strings.Contains(str, expectedStr), nil
		}
	}

	// Handle slice contains
	if slice, ok := value.([]interface{}); ok {
		for _, item := range slice {
			matches, err := evaluateEq(item, expected)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle []string
	if strSlice, ok := value.([]string); ok {
		expectedStr, ok := expected.(string)
		if !ok {
			return false, nil
		}
		for _, s := range strSlice {
			if s == expectedStr {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}

// evaluateStartsWith evaluates string starts with.
func evaluateStartsWith(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	expectedStr, ok := expected.(string)
	if !ok {
		return false, nil
	}
	return strings.HasPrefix(str, expectedStr), nil
}

// evaluateEndsWith evaluates string ends with.
func evaluateEndsWith(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	expectedStr, ok := expected.(string)
	if !ok {
		return false, nil
	}
	return strings.HasSuffix(str, expectedStr), nil
}

// evaluateMatches evaluates regex match.
func evaluateMatches(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	pattern, ok := expected.(string)
	if !ok {
		return false, nil
	}
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}
	return matched, nil
}

// evaluateDefined evaluates if a property is defined.
func evaluateDefined(value interface{}, expected interface{}) (bool, error) {
	defined, ok := expected.(bool)
	if !ok {
		return false, nil
	}
	if defined {
		// Check if value is defined (not nil and not empty string)
		if value == nil {
			return false, nil
		}
		if str, ok := value.(string); ok && str == "" {
			return false, nil
		}
		return true, nil
	}
	// Check if value is undefined (nil or empty string)
	if value == nil {
		return true, nil
	}
	if str, ok := value.(string); ok && str == "" {
		return true, nil
	}
	return false, nil
}
