package transformjson

import (
	"lazy-lagoon/pkg/types"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestExpressions(t *testing.T) {
	jsonDocument := FileToJson(t, "../assets/goldenFiles/test.json")
	// Test Case 1: Simple name matching
	t.Run("Simple name matching", func(t *testing.T) {
		expression := types.Expression{
			Expressions: []types.Expressions{
				{
					FieldName: "friends[*].name",
					Operator:  "EQ",
					Value:     "Bob",
				},
			},
			LogicalOperator: "AND",
		}

		// Checking friends.0.name is false
		met, err := IsExpressionMet(expression, []int{0}, 0, jsonDocument)
		if err != nil {
			t.Fatalf("Failed to check if expression is met: %v", err)
		}
		assert.Equal(t, met, false)

		// Checking friends.1.name is true
		met, err = IsExpressionMet(expression, []int{1}, 0, jsonDocument)
		if err != nil {
			t.Fatalf("Failed to check if expression is met: %v", err)
		}
		assert.Equal(t, met, true)
	})

	// Test Case 2: Complex nested array matching
	t.Run("complex nested array matching", func(t *testing.T) {
		expression := types.Expression{
			Expressions: []types.Expressions{
				{
					FieldName: "friends[*].contacts[*].type",
					Operator:  "EQ",
					Value:     "email",
				},
			},
			LogicalOperator: "AND",
		}

		// Checking friends.0.contacts.0.type is true
		met, err := IsExpressionMet(expression, []int{0, 0}, 0, jsonDocument)
		if err != nil {
			t.Fatalf("Failed to check if expression is met: %v", err)
		}
		assert.Equal(t, met, true)

		// Checking friends.1.contacts.0.type is false		
		met, err = IsExpressionMet(expression, []int{1, 0}, 0, jsonDocument)
		if err != nil {
			t.Fatalf("Failed to check if expression is met: %v", err)
		}
		assert.Equal(t, met, false)

		// Checking friends.1.contacts.1.type is true
		met, err = IsExpressionMet(expression, []int{1, 1}, 0, jsonDocument)
		if err != nil {
			t.Fatalf("Failed to check if expression is met: %v", err)
		}
		assert.Equal(t, met, true)
	})
}
