package types

type TransformError struct {
	Message         string `json:"message"`
	RuleIndex       *int   `json:"ruleIndex,omitempty"`
	ActionIndex     *int   `json:"actionIndex,omitempty"`
	ExpressionIndex *int   `json:"expressionIndex,omitempty"`
	Key             string `json:"key,omitempty"`
}

type Action struct {
	ActionType string `json:"actionType"`
	FieldName  string `json:"fieldName"`
}

type Expression struct {
	LogicalOperator string        `json:"logicalOperator"`
	Expressions     []Expressions `json:"expressions"`
}

type Expressions struct {
	FieldName string      `json:"fieldName"`
	Operator  string      `json:"operator"`
	Value     any 				`json:"value"`
}

type Rule struct {
	Expression Expression `json:"expression,omitempty"`
	Actions    []Action   `json:"actions"`
}
