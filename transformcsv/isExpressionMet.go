package transformcsv

import (
	"lazy-lagoon/pkg/expressions"
	"lazy-lagoon/pkg/types"
	"slices"
)

/*
	IsExpressionMet checks if the given expression is met
*/
func IsExpressionMet(expression types.Expression, lines [][]string, index int, ruleIndex int) (bool, *types.TransformError) {
	expressionMet := false
	if expression.Expressions == nil {
		return true, nil
	}
	for expressionIndex, exp := range expression.Expressions {
		// Checking to see if the expression is met
		expressionColumn := slices.Index(lines[0], exp.FieldName)
		if expressionColumn == -1 {
			// Column not found but considering it a no op
			return false, nil
		}
		if expressionColumn >= len(lines[0]) {
			// expression header is out of range
			return false, &types.TransformError{
				Message: "expression header is out of range",
				RuleIndex: &ruleIndex,
				ExpressionIndex: &expressionIndex,
				Key: "fieldName",
			}
		}
		// Dealing with logical operators
		met, err := expressions.IsOperatorResultMet(exp.Operator, exp.Value, lines[index][expressionColumn])
		if err != nil {
			return false, &types.TransformError{
				Message: err.Error(),
				RuleIndex: &ruleIndex,
				ExpressionIndex: &expressionIndex,
				Key: "value",
			}
		}
		// Dealing with logical operators
		if met {
			// If the logical operator is or then return true
			if expression.LogicalOperator == "OR" {
				return true, nil
			} else {
				// If the logical operator is and then set the expressionMet to true (we have to check all expressions are met)
				expressionMet = true
			}
		} else {
			// If the logical operator is and then return false
			if expression.LogicalOperator == "AND" {
				return false, nil
			}
		}
	}

	return expressionMet, nil
}
