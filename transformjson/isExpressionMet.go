package transformjson

import (
	"lazy-lagoon/pkg/expressions"
	"lazy-lagoon/pkg/types"
	"strconv"
	"strings"
)

/*
IsExpressionMet checks if the given expression is met
*/
func IsExpressionMet(expression types.Expression, indexes []int, ruleIndex int, jsonDocument any) (bool, *types.TransformError) {
	// If the expression is nil then return true
	if expression.Expressions == nil {
		return true, nil
	}

	// For AND, start optimistic; for OR start pessimistic
	expressionMet := expression.LogicalOperator != "OR"
	isOperatorResultMet := false

	for expressionIndex, exp := range expression.Expressions {
		// Process the field name with indexes if needed
		fieldName := exp.FieldName
		if fieldName == "" {
			// SKIP the expression if the field name is empty
			return false, nil
		}
		if len(indexes) > 0 {
			fieldName = replaceIndexes(fieldName, indexes)
		}

		// If the fieldName does not contain a wildcard.
		if !strings.Contains(fieldName, "*") {
			/*
				Checking the expression on the single value of the field.
			*/
			// Get the value from the JSON document
			expressionValue, err := getExpressionValue(fieldName, jsonDocument)
			if err != nil {
				return false, &types.TransformError{
					Message:         err.Error(),
					RuleIndex:       &ruleIndex,
					ExpressionIndex: &expressionIndex,
					Key:             "fieldName",
				}
			}

			// Check if the operator condition is met
			isOperatorResultMet, err = expressions.IsOperatorResultMet(exp.Operator, exp.Value, expressionValue)
			if err != nil {
				return false, &types.TransformError{
					Message:         err.Error(),
					RuleIndex:       &ruleIndex,
					ExpressionIndex: &expressionIndex,
					Key:             "comparator",
				}
			}
		} else {
			/*
				Checking the array of values for the case where an expression has extra arrays, if any of the values meet the expression then return true. (Logical OR)
			*/
			// Get the values from the JSON document
			expressionArrayValues, err := getExpressionArrayValues(fieldName, jsonDocument)
			if err != nil {
				return false, &types.TransformError{
					Message:         err.Error(),
					RuleIndex:       &ruleIndex,
					ExpressionIndex: &expressionIndex,
					Key:             "fieldName",
				}
			}

			// Check if the operator condition is met
			isOperatorResultMet, err = expressions.IsOperatorArrayResultMet(exp.Operator, exp.Value, expressionArrayValues)
			if err != nil {
				return false, &types.TransformError{
					Message:         err.Error(),
					RuleIndex:       &ruleIndex,
					ExpressionIndex: &expressionIndex,
					Key:             "comparator",
				}
			}
		}

		// Handle logical operators
		if isOperatorResultMet {
			// For OR operations, return true immediately if any condition is met
			if expression.LogicalOperator == "OR" {
				return true, nil
			}
			// For AND operations, mark that at least one condition is met
			expressionMet = true
		} else {
			// For AND operations, return false immediately if any condition is not met
			if expression.LogicalOperator == "AND" {
				return false, nil
			}
		}
	}

	return expressionMet, nil
}

// replaceIndexes replaces all * characters in the field name with the corresponding index
func replaceIndexes(fieldName string, indexes []int) string {
	result := fieldName
	for _, index := range indexes {
		result = strings.Replace(result, "*", strconv.Itoa(index), 1)
	}
	return result
}

// getExpressionArrayValues gets the values from the JSON document using the field name
func getExpressionArrayValues(fieldName string, jsonDocument any) ([]any, error) {
	// Make a pointer to the field
	expressionPointer, err := MakePointer(fieldName)
	if err != nil {
		return nil, err
	}

	// Get the value from the pointer
	expressionValues, err := GetPointerArrayValues(expressionPointer, jsonDocument)
	if err != nil {
		return nil, err
	}

	return expressionValues, nil
}

// getExpressionValue gets the value from the JSON document using the field name
func getExpressionValue(fieldName string, jsonDocument any) (any, error) {
	// Make a pointer to the field
	expressionPointer, err := MakePointer(fieldName)
	if err != nil {
		return nil, err
	}

	// Get the value from the pointer
	expressionValue, err := GetPointerValue(expressionPointer, jsonDocument)
	if err != nil {
		return nil, err
	}

	return expressionValue, nil
}
