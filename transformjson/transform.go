package transformjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lazy-lagoon/pkg/types"
	"lazy-lagoon/storage"
)

/*
Step 1: execute transform, and store in output storage
*/
func ExecuteTransform(input types.Input, rules []types.Rule, output types.Output) *types.TransformError {
	/*
		Downloading the file from the input storage type
	*/
	byteContent, err := storage.GetBytes(input)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}
	/*
		Parsing the JSON document
	*/
	jsonDocument, err := ToJson(byteContent)
	if err != nil {
		return &types.TransformError{
			Message: err.Error(),
		}
	}
	/*
		Transforming the JSON document
	*/
	jsonDocument, transformErr := ExecuteRules(jsonDocument, rules)
	if transformErr != nil {
		return transformErr
	}
	/*
		Converting the JSON document back to bytes
	*/
	byteContent, err = FromJson(jsonDocument)
	if err != nil {
		return &types.TransformError{
			Message: err.Error(),
		}
	}

	// Store the transformed JSON bytes in the output storage type
	err = storage.StoreBytes(output, byteContent)
	if err != nil {
		return &types.TransformError{
			Message: err.Error(),
		}
	}
	return nil
}

/*
Step 2: Transform the JSON document based on the rules
*/
func ExecuteRules(jsonDocument any, rules []types.Rule) (any, *types.TransformError) {
	// Apply the rules to the document
	for ruleIndex, rule := range rules {
		for actionIndex, action := range rule.Actions {
			var transformErr *types.TransformError
			jsonDocument, transformErr = ExecuteAction(jsonDocument, rule.Expression, action.FieldName, action.ActionType, ruleIndex, actionIndex)
			if transformErr != nil {
				return nil, transformErr
			}
		}
	}
	return jsonDocument, nil
}

/*
Step 3: Execute the actions by the actionType if the expressions are met
*/
func ExecuteAction(jsonDocument any, expression types.Expression, fieldName string, actionType string, ruleIndex int, actionIndex int) (any, *types.TransformError) {
	if fieldName == "" {
		// SKIP the action if the field name is empty
		return jsonDocument, nil
	}
	// Creating a json pointer
	pointer, err := MakePointer(fieldName)
	if err != nil {
		return nil, &types.TransformError{
			Message:     err.Error(),
			RuleIndex:   &ruleIndex,
			ActionIndex: &actionIndex,
			Key:         "fieldName",
		}
	}

	// Creating a copy of the json document so the original is not mutated
	var documentCopy any
	if err := DeepCopyJSON(jsonDocument, &documentCopy); err != nil {
		return nil, &types.TransformError{
			Message:     err.Error(),
			RuleIndex:   &ruleIndex,
			ActionIndex: &actionIndex,
			Key:         "fieldName",
		}
	}

	// Manipulating the json document from the pointer tokens
	transformErr := Mutate(documentCopy, documentCopy, pointer, expression, actionType, []int{}, ruleIndex, actionIndex)
	if transformErr != nil {
		return nil, transformErr
	}

	// Returning the mutated document
	return documentCopy, nil
}

/*
	Helper functions
*/
// Deep copy the JSON document
func DeepCopyJSON(src, dst any) error {
	// 1) Marshal the source into JSON bytes
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	// 2) Unmarshal into the destination (which must be a pointer)
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}
	return nil
}

// Convert the bytes to a JSON document
func ToJson(content []byte) (any, error) {
	var jsonDocument any
	// Unmarshal root json
	err := json.Unmarshal(content, &jsonDocument)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json error, not valid json")
	}
	return jsonDocument, nil
}

// Convert the JSON document to bytes
func FromJson(jsonDocument any) ([]byte, error) {
	// marshalling json back together
	newBytes, err := json.Marshal(jsonDocument)
	if err != nil {
		return nil, err
	}
	
	// Fixing json indentation to be pretty :)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, newBytes, "", "\t")
	if err != nil {
		return nil, err
	}

	return prettyJSON.Bytes(), nil
}

// Convert the JSONL document to bytes - basically just not prettifying the json
func FromJsonl(jsonDocument any) ([]byte, error) {
	// marshalling json back together
	newBytes, err := json.Marshal(jsonDocument)
	if err != nil {
		return nil, err
	}
	
	return newBytes, nil
}