package transformjson

import (
	"fmt"
	"lazy-lagoon/pkg/types"
	"strconv"
	"strings"
)

/*
MakePointer splits token string by . to individual tokens and handles array indices
*/
func MakePointer(token string) ([]string, error) {
	if token == "" {
		return nil, fmt.Errorf("no field provided")
	}

	var tokens []string
	parts := strings.Split(token, ".")

	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty token in path")
		}

		// Check if this part contains array notation
		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			// Split into array name and index
			arrayParts := strings.Split(part, "[")
			if len(arrayParts) != 2 {
				return nil, fmt.Errorf("invalid array notation in %s", part)
			}

			// Add the array name as a token
			if arrayParts[0] != "" {
				tokens = append(tokens, arrayParts[0])
			}

			// Add the index (including *) as a token
			index := strings.TrimSuffix(arrayParts[1], "]")
			tokens = append(tokens, index)
		} else {
			tokens = append(tokens, part)
		}
	}

	return tokens, nil
}

/*
Get the value from the json pointer
*/
func GetPointerValue(tokens []string, node any) (any, error) {
	if len(tokens) == 0 {
		return "", fmt.Errorf("empty token list")
	}

	currentToken := tokens[0]
	isLastToken := len(tokens) == 1
	cleanedToken := tokens[1:]

	// Reached the end of the pointer
	if isLastToken {
		switch typedNode := node.(type) {
		// If the node is a object
		case map[string]any:
			// If the node is a map, try to get the value from the map
			if value, exists := typedNode[currentToken]; exists {
				return fmt.Sprintf("%v", value), nil
			}
			return "", nil
		case []any:
			// If the node is an array, try to convert the current token to an integer and index into the array
			currentTokenAsInt, err := strconv.Atoi(currentToken)
			if err != nil {
				return "", fmt.Errorf("invalid array index format: %v", err)
			}
			if currentTokenAsInt < 0 || currentTokenAsInt >= len(typedNode) {
				return "", fmt.Errorf("array index out of bounds: %d (length: %d)", currentTokenAsInt, len(typedNode))
			}
			return typedNode[currentTokenAsInt], nil
		case any, nil:
			return typedNode, nil
		default:
			return "", fmt.Errorf("invalid node type for getting value")
		}
	}
	switch typedNode := node.(type) {
	case map[string]any:
		// If the node is a map then index into the map and recurse
		if value, ok := typedNode[currentToken]; ok {
			return GetPointerValue(cleanedToken, value)
		}
		return "", nil
	case []any:
		// If the current token is a wildcard then throw an error (should be using GetPointerArrayValues)
		if currentToken == "*" {
			return "", fmt.Errorf("cannot use * as an index")
		}
		// If the current token is an integer then index into the array and recurse
		if index, err := strconv.Atoi(currentToken); err == nil {
			if index >= 0 && index < len(typedNode) {
				return GetPointerValue(cleanedToken, typedNode[index])
			}
			return "", fmt.Errorf("invalid array index")
		}
	}
	return node, nil
}

/*
Get multiple values from the json pointer
*/
func GetPointerArrayValues(tokens []string, node any) ([]any, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty token list")
	}
	currentToken := tokens[0]
	isLastToken := len(tokens) == 1

	if isLastToken {
		// Handle potential errors based on node type
		switch typedNode := node.(type) {
		// If the node is a object
		case map[string]any:
			// if the node is a map and current token is a key then return the value as an array
			if value, exists := typedNode[currentToken]; exists {
				return []any{value}, nil
			}
			return nil, nil
		case []any:
			// Return the array directly since it's already []any
			return typedNode, nil
		case any:
			return []any{typedNode}, nil
		}
	}

	cleanedToken := tokens[1:]
	switch typedNode := node.(type) {
	// If the node is a object
	case map[string]any:
		// If the node is a map then index into the map and recurse
		if value, ok := typedNode[currentToken]; ok {
			return GetPointerArrayValues(cleanedToken, value)
		}
		return nil, nil
	// If the node is an array
	case []any:
		if currentToken == "*" {
			var results []any
			// Getting all the values for each array *
			for _, child := range typedNode {
				childValues, err := GetPointerArrayValues(cleanedToken, child)
				if err != nil {
					return nil, err
				}
				results = append(results, childValues...)
			}
			return results, nil
		} else if index, err := strconv.Atoi(currentToken); err == nil {
			// If the current token is an integer then index into the array and recurse
			if index >= 0 && index < len(typedNode) {
				return GetPointerArrayValues(cleanedToken, typedNode[index])
			}
			return nil, fmt.Errorf("invalid array index")
		}
	}
	return []any{node}, nil
}

/*
Mutate goes through node interface from the given pointer to Redact or Exclude values in json and checks if the expression is met
*/
func Mutate(
	// Document is getting compared and reviewed for the expression and then mutated if the expression is met
	// IMPORTANT: This function mutates the document
	document any,
	// Node is the current node that is being evaluated. Its recursively indexed through each token.
	// WARNING: This variable gets mutated as we traverse through the pointer.
	node any,
	// Tokens are the tokens that are being traversed through the pointer
	tokens []string,
	expression types.Expression,
	actionType string,
	// Indexes is a collection of indexes that have been traversed through the pointer. Its used to keep track of max indexes to check in the expression.
	indexes []int,
	ruleIndex int,
	actionIndex int,
	) *types.TransformError {
	if len(tokens) == 0 {
		return &types.TransformError{
			Message: "empty token list",
			RuleIndex: &ruleIndex,
			ActionIndex: &actionIndex,
			Key: "fieldName",
		}
	}
	currentToken := tokens[0]
	isLastToken := len(tokens) == 1
	cleanedToken := tokens[1:]

	if isLastToken {
		// At the last token, we may be redacting/excluding a value or array element
		met, transformErr := IsExpressionMet(expression, indexes, ruleIndex, document)
		if transformErr != nil {
			return transformErr
		}
		if !met {
			// Returning without redaction if expression isn't met
			return nil
		}
		switch typedNode := node.(type) {
		// If the node is a map/object
		case map[string]any:
			if actionType == "REDACT" {
				// Checking to see if the value exists
				if _, exists := typedNode[currentToken]; exists {
					// Redact the value
					typedNode[currentToken] = "**redacted**"
				}
			} else if actionType == "EXCLUDE" {
				// Exclude (delete) the key
				delete(typedNode, currentToken)
			}
			return nil
		// If the node is an array
		case []any:
			// if the last token is * and an array
			if currentToken == "*" {
				for i := 0; i < len(typedNode); i++ {
					// Calling again so individually can check expressions
					if transformErr := Mutate(document, typedNode, []string{strconv.Itoa(i)}, expression, actionType, append(indexes, i), ruleIndex, actionIndex); transformErr != nil {
						return transformErr
					}
				}
			} else if currentToken != "" {
				// The * after its been called above. (execution phase)
				tokenAsInt, err := strconv.Atoi(currentToken)
				if err != nil {
					return &types.TransformError{
						Message: fmt.Sprintf("invalid array index format: %v", err),
						RuleIndex: &ruleIndex,
						ActionIndex: &actionIndex,
						Key: "fieldName",
					}
				}
				if actionType == "REDACT" {
					typedNode[tokenAsInt] = "**redacted**"
				} else if actionType == "EXCLUDE" {
					typedNode[tokenAsInt] = []interface{}{}
				}
			}
			return nil
		default:
			// Skipping if the node type is not supported
			return nil
		}
	}

	// Not at the last token: recurse deeper into the structure
	switch typedNode := node.(type) {
	// If the node is a map/object
	case map[string]any:
		if value, ok := typedNode[currentToken]; ok {
			// Recurse into the next token
			return Mutate(document, value, cleanedToken, expression, actionType, indexes, ruleIndex, actionIndex)
		} else {
			// No op if the key doesn't exist in this index
			return nil
		}
	// If the node is an array
	case []any:
		if currentToken == "*" {
			// Wildcard: recurse into each child with the current index
			for index, child := range typedNode {
				if transformErr := Mutate(document, child, cleanedToken, expression, actionType, append(indexes, index), ruleIndex, actionIndex); transformErr != nil {
					return transformErr
				}
			}
			return nil
		} else if currentToken != "" {
			// Handle a specific array index
			tokenAsInt, err := strconv.Atoi(currentToken)
			if err != nil {
				return &types.TransformError{
					Message: fmt.Sprintf("invalid array index format: %v", err),
					RuleIndex: &ruleIndex,
					ActionIndex: &actionIndex,
					Key: "fieldName",
				}
			}
			if tokenAsInt < 0 || tokenAsInt >= len(typedNode) {
				return &types.TransformError{
					Message: fmt.Sprintf("array index out of bounds: %d (length: %d)", tokenAsInt, len(typedNode)),
					RuleIndex: &ruleIndex,
					ActionIndex: &actionIndex,
					Key: "fieldName",
				}
			}
			// Recurse into the next token for the given index
			return Mutate(document, typedNode[tokenAsInt], cleanedToken, expression, actionType, append(indexes, tokenAsInt), ruleIndex, actionIndex)
		}
		return nil
	}
	// Skipping if the node type is not supported
	return nil
}
