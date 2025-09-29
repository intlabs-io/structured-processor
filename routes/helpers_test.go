package routes

import (
	"testing"
	
	"github.com/go-playground/assert/v2"
)

func EqualNotSorted(t *testing.T, expected, actual []string) {
	assert.Equal(t, len(expected), len(actual))
	for _, exp := range expected {
		assert.Equal(t, true, contains(actual, exp))
	}
}

func TestExtractPaths(t *testing.T) {
	t.Run("JSON simple objects", func(t *testing.T) {
		jsonData := []byte(`{
			"name": "John",
			"age": 30,
			"address": {
				"street": "123 Main St",
				"city": "Boston"
			}
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"name",
			"age",
			"address",
			"address.street",
			"address.city",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON arrays of objects", func(t *testing.T) {
		jsonData := []byte(`{
			"users": [
				{ "id": 1, "name": "John" },
				{ "id": 2, "name": "Jane" }
			]
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"users", 
			"users[*]", 
			"users[*].id", 
			"users[*].name",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON nested arrays", func(t *testing.T) {
		jsonData := []byte(`{
			"departments": [
				{
					"name": "Engineering",
					"teams": [
						{ "id": 1, "name": "Frontend" },
						{ "id": 2, "name": "Backend" }
					]
				}
			]
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"departments",
			"departments[*]",
			"departments[*].name",
			"departments[*].teams",
			"departments[*].teams[*]",
			"departments[*].teams[*].id",
			"departments[*].teams[*].name",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON empty arrays", func(t *testing.T) {
		jsonData := []byte(`{
			"emptyArray": []
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"emptyArray", 
			"emptyArray[*]",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON primitive values", func(t *testing.T) {
		jsonData := []byte(`{
			"string": "hello",
			"number": 42,
			"boolean": true,
			"null": null
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"string", 
			"number", 
			"boolean", 
			"null",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON varying properties across elements", func(t *testing.T) {
		jsonData := []byte(`{
			"users": [
				{ "id": 2, "name": "Jane", "email": "jane@example.com" },
				{ "id": 1, "name": "John", "age": 30, "roles": ["admin"] },
				{ "id": 3, "name": "Bob", "department": { "id": 5, "name": "Engineering" } }
			],
			"metadata": {
				"version": "1.0",
				"settings": {
					"isPublic": true,
					"permissions": ["read", "write"]
				}
			},
			"stats": [
				{ "date": "2023-01-01", "count": 10, "details": { "success": 8 } },
				{ "date": "2023-01-02", "count": 15 }
			]
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"users",
			"users[*]",
			"users[*].id",
			"users[*].name",
			"users[*].email",
			"users[*].age",
			"users[*].roles",
			"users[*].roles[*]",
			"users[*].department",
			"users[*].department.id",
			"users[*].department.name",
			"metadata",
			"metadata.version",
			"metadata.settings",
			"metadata.settings.isPublic",
			"metadata.settings.permissions",
			"metadata.settings.permissions[*]",
			"stats",
			"stats[*]",
			"stats[*].date",
			"stats[*].count",
			"stats[*].details",
			"stats[*].details.success",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSON 2D arrays", func(t *testing.T) {
		jsonData := []byte(`{
			"matrix": [
				[1, 2],
				[3, 4]
			],
			"nestedData": {
				"grid": [[{ "value": "a" }], [{ "value": "b" }]]
			}
		}`)
		paths, err := extractPaths(jsonData, "JSON")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"matrix",
			"matrix[*]",
			"matrix[*][*]",
			"nestedData",
			"nestedData.grid",
			"nestedData.grid[*]",
			"nestedData.grid[*][*]",
			"nestedData.grid[*][*].value",
		}
		
		EqualNotSorted(t, expected, paths)
	})

	t.Run("JSONL", func(t *testing.T) {
		jsonlData := []byte(`{"name": "Alice", "age": 30, "address": {"city": "New York", "zip": "10001"}}
{"name": "Bob", "age": 25, "hobbies": ["reading", "gaming"]}
{"name": "Charlie", "email": "charlie@example.com", "metadata": {"joined": "2023-01-15", "status": "active"}}`)
		paths, err := extractPaths(jsonlData, "JSONL")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"name",
			"age",
			"address",
			"address.city",
			"address.zip",
			"hobbies",
			"hobbies[*]",
			"email",
			"metadata",
			"metadata.joined",
			"metadata.status",
		}
		EqualNotSorted(t, expected, paths)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"name": "Alice", "age": }`)
		_, err := extractPaths(invalidJSON, "JSON")
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
	
	t.Run("CSV", func(t *testing.T) {
		csvData := []byte("name,age,city\nAlice,30,Wonderland\nBob,25,Atlantis\n")
		paths, err := extractPaths(csvData, "CSV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"name", "age", "city"}
		EqualNotSorted(t, expected, paths)
	})

}
