package transformjson

import (
	"testing"

	"lazy-lagoon/pkg/types"
	"os"

	"github.com/go-playground/assert/v2"
)

func FileToJson(t *testing.T, filename string) any {
	originalBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	json, err := ToJson(originalBytes)
	if err != nil {
		t.Fatalf("Failed to convert to json: %v", err)
	}
	return json
}

func TestConversionJson(t *testing.T) {
	json := FileToJson(t, "../assets/goldenFiles/testUnmarshalling.json")
	
	// Verify the JSON was loaded correctly
	if json == nil {
		t.Fatalf("Failed to load JSON: got nil")
	}
	
	// Test that we can convert it back to bytes
	roundTrippedBytes, err := FromJson(json)
	if err != nil {
		t.Fatalf("Failed to convert JSON back to bytes: %v", err)
	}

	roundTrippedJson, err := ToJson(roundTrippedBytes)
	if err != nil {
		t.Fatalf("Failed to convert bytes back to JSON: %v", err)
	}

	// Now we can compare the original and round-tripped JSON objects
	assert.Equal(t, json, roundTrippedJson)
}

func TestRuleNoExpression(t *testing.T) {
	originalJson := FileToJson(t, "../assets/goldenFiles/test.json")
	t.Run("1. no expression redact array * obj * value", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesNoExpression1Redacted.json")

		expression := types.Expression{}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].contacts[*].value",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("2. no expression redact array *", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesNoExpression2Redacted.json")

		expression := types.Expression{}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].contacts[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("3. no expression exclude array * obj * value", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesNoExpression3Exclude.json")

		expression := types.Expression{}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].contacts[*].value",
						ActionType: "EXCLUDE",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})
}

func TestRuleExpression(t *testing.T) {
	originalJson := FileToJson(t, "../assets/goldenFiles/test.json")

	t.Run("1. equal arrays in rules and expressions", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesExpression1Redacted.json")

		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "friends[*].contacts[*].value",
					Operator:  "EQ",
					Value:     "alice@example.com",
				},
			},
		}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].contacts[*].value",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("2. more arrays in expressions than action", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesExpression2Redacted.json")

		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "friends[*].contacts[*].value",
					Operator:  "EQ",
					Value:     "alice@example.com",
				},
			},
		}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].name",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("3. more arrays in action than expressions", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesExpression3Redacted.json")

		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "friends[*].name",
					Operator:  "EQ",
					Value:     "Alice",
				},
			},
		}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "friends[*].contacts[*].value",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("4. string array with EQ operator", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesExpression4Redacted.json")

		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "roles[*]",
					Operator:  "EQ",
					Value:     "guest",
				},
			},
		}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "city",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("5. redacting specific value of array item", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testRulesExpression5Redacted.json")

		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "roles[*]",
					Operator:  "EQ",
					Value:     "admin",
				},
			},
		}

		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "roles[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})
}

func TestObscure(t *testing.T) {
	originalJson := FileToJson(t, "../assets/goldenFiles/testObscure.json")
	t.Run("redact array", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure1Redacted.json")
		
		expression := types.Expression{}
		
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "[*].array",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("redact array elements with expression", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure2Redacted.json")
		
		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "[*].array[*]",
					Operator:  "EQ",
					Value:     "test2",
				},
			},
		}
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "[*].array[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("redact int type", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure3Redacted.json")
		
		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "[*]",
					Operator:  "EQ",
					Value:     "1",
				},
			},
		}
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("redact null type", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure4Redacted.json")
		
		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "[*]",
					Operator:  "EQ",
					Value:     "null",
				},
			},
		}
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("redact array with wildcard", func(t *testing.T) {
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure5Redacted.json")
		
		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "[*].array[*]",
					Operator:  "EQ",
					Value:     "test2",
				},
			},
		}
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "[*]",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})

	t.Run("redact array with EXISTS operator", func(t *testing.T) {
		originalJson := FileToJson(t, "../assets/goldenFiles/testObscure6.json")
		redactedJson := FileToJson(t, "../assets/goldenFiles/testObscure6Redacted.json")
		
		expression := types.Expression{
			LogicalOperator: "AND",
			Expressions: []types.Expressions{
				{
					FieldName: "array[*].test2",
					Operator:  "EXISTS",
					Value:     true,
				},
			},
		}
		rules := []types.Rule{
			{
				Expression: expression,
				Actions: []types.Action{
					{
						FieldName:  "array[*].test",
						ActionType: "REDACT",
					},
				},
			},
		}

		mutatedJson, err := ExecuteRules(originalJson, rules)
		if err != nil {
			t.Fatalf("Failed to execute rule: %v", err)
		}

		assert.Equal(t, mutatedJson, redactedJson)
	})
}
