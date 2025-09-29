package expressions

import (
	"fmt"
	"strconv"
)

// ValidOperators contains the allowed operator values
var ValidOperators = []string{"EQ", "NE", "GT", "GTE", "LT", "LTE", "EXISTS"}

// IsValidOperator checks if the given operator is valid
func IsValidOperator(operator string) bool {
	for _, validOp := range ValidOperators {
		if operator == validOp {
			return true
		}
	}
	return false
}

/*
	IsOperatorMet checks if the given expression params match the given operator result
*/
func IsOperatorResultMet(operator string, expectedValue any, actualValue any) (bool, error) {
	if !IsValidOperator(operator) {
		return false, fmt.Errorf("invalid operator: %s", operator)
	}

	canConvertToFloat := true

	actualString := fmt.Sprintf("%v", actualValue)
	expectedString := fmt.Sprintf("%v", expectedValue)
	actualFloat, err := strconv.ParseFloat(actualString, 64)
	if err != nil {
		canConvertToFloat = false
	}
	expectedFloat, err := strconv.ParseFloat(expectedString, 64)
	if err != nil {
		canConvertToFloat = false
	}

	switch operator {
	case "EQ":
		// Handle null values
		if expectedValue == "null" && actualValue == nil {
			return true, nil
		}
		// Handle string comparison
		return actualString == expectedString, nil
	case "NE":
		// Handle null values
		if expectedValue == "null" && actualValue != nil {
			return true, nil
		}
		// Handle string comparison
		return actualString != expectedString, nil
	case "GT":
		// Handle numeric comparison
		if canConvertToFloat {
			return actualFloat > expectedFloat, nil
		}
		// Handle string comparison
		return actualString > expectedString, nil
	case "GTE":
		if canConvertToFloat {
			return actualFloat >= expectedFloat, nil
		}
		return actualString >= expectedString, nil
	case "LT":
		if canConvertToFloat {
			return actualFloat < expectedFloat, nil
		}
		return actualString < expectedString, nil
	case "LTE":
		if canConvertToFloat {
			return actualFloat <= expectedFloat, nil
		}
		return actualString <= expectedString, nil
	case "EXISTS":
		if (expectedValue == true) {
			// EXISTS
			return actualString != "", nil
		} else {
			// NOT EXISTS
			return actualString == "", nil
		}
	}

	return false, nil
}

/*
	IsOperatorArrayResultMet looping through the array of values and checking if any of the values meet the condition (Logical OR)
*/
func IsOperatorArrayResultMet(operator string, expectedValue any, actualValue []any) (bool, error) {
	if !IsValidOperator(operator) {
		return false, fmt.Errorf("invalid operator: %s", operator)
	}

	// If there are no actual values, the condition can't be met
	if len(actualValue) == 0 {
		return false, nil
	}

	// For each value in the array, check if it meets the condition
	for _, value := range actualValue {
		// Use the existing IsOperatorResultMet function to evaluate each value
		result, err := IsOperatorResultMet(operator, expectedValue, value)
		if err != nil {
			return false, err
		}
		// If any of the values meet the condition, return true
		if result {
			return true, nil
		}
	}

	// If no values matched the condition
	return false, nil

}
