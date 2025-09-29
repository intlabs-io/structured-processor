package transformcsv

import (
	"testing"

	"lazy-lagoon/pkg/types"
	"os"

	"github.com/go-playground/assert/v2"
)

func FileToCsv(t *testing.T, filename string) [][]string {
	originalBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	csv, err := ToCsv(originalBytes)
	if err != nil {
		t.Fatalf("Failed to convert to csv: %v", err)
	}
	return csv
}

func TestRulesNoExpression(t *testing.T) {
	originalCsv := FileToCsv(t, "../assets/goldenFiles/testRulesNoExpression.csv")
	t.Run("redact csv field", func(t *testing.T) {
		// Setup test data
		redactedCsv := FileToCsv(t, "../assets/goldenFiles/testRulesNoExpression1Redacted.csv")

		// Create rule with empty expression and multiple actions
		rules := []types.Rule{{
			Expression: types.Expression{},
			Actions: []types.Action{
				{ActionType: "REDACT", FieldName: "First name"},
				{ActionType: "REDACT", FieldName: "Last name"},
			},
		}}

		// Execute rules
		mutatedCsv, err := ExecuteRules(originalCsv, rules)
		if err != nil {
			t.Fatalf("Failed to execute rules: %v", err)
		}

		assert.Equal(t, mutatedCsv, redactedCsv)
	})
}

func TestRulesEmptyColumns(t *testing.T) {
	originalCsv := FileToCsv(t, "../assets/goldenFiles/testRulesEmptyColumns.csv")
	t.Run("1. redact empty columns", func(t *testing.T) {
		// Setup test data
		redactedCsv := FileToCsv(t, "../assets/goldenFiles/testRulesEmptyColumns1Redacted.csv")

		// Create rule with empty expression and multiple actions
		rules := []types.Rule{{
			Expression: types.Expression{},
			Actions: []types.Action{
				{ActionType: "REDACT", FieldName: "Email"},
				{ActionType: "REDACT", FieldName: "Last Name"},
				{ActionType: "REDACT", FieldName: "First Name"},
			},
		}}

		// Execute rules
		mutatedCsv, err := ExecuteRules(originalCsv, rules)
		if err != nil {
			t.Fatalf("Failed to execute rules: %v", err)
		}

		assert.Equal(t, mutatedCsv, redactedCsv)
	})
}

func TestRulesExpression(t *testing.T) {
	originalCsv := FileToCsv(t, "../assets/goldenFiles/testRulesExpression.csv")

	t.Run("1. redact with expression", func(t *testing.T) {
		// Setup test data
		redactedCsv := FileToCsv(t, "../assets/goldenFiles/testRulesExpression1Redacted.csv")

		// Create rule with payment expression
		rules := []types.Rule{{
			Expression: types.Expression{
				Expressions: []types.Expressions{
					{FieldName: "type", Operator: "EQ", Value: "PAYMENT"},
				},
			},
			Actions: []types.Action{
				{ActionType: "REDACT", FieldName: "type"},
				{ActionType: "REDACT", FieldName: "amount"},
			},
		}}

		// Execute rules
		mutatedCsv, err := ExecuteRules(originalCsv, rules)
		if err != nil {
			t.Fatalf("Failed to execute rules: %v", err)
		}

		assert.Equal(t, mutatedCsv, redactedCsv)
	})
}
