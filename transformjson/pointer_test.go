package transformjson

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMakeAndGetPointer(t *testing.T) {
	originalJson := FileToJson(t, "../assets/goldenFiles/test.json")

	t.Run("multi array pointer", func(t *testing.T) {
		friendsMultiNamePointer, err := MakePointer("friends[*].name")
		if err != nil {
			t.Fatalf("Failed to create pointer: %v", err)
		}
		friendsMultiNameValues, err := GetPointerArrayValues(friendsMultiNamePointer, originalJson)
		if err != nil {
			t.Fatalf("Failed to get pointer value: %v", err)
		}
		assert.Equal(t, friendsMultiNameValues, []any{"Alice", "Bob"})
	})

	t.Run("single array pointer", func(t *testing.T) {
		friendsSingleNamePointer, err := MakePointer("friends[1].name")
		if err != nil {
			t.Fatalf("Failed to create pointer: %v", err)
		}

		friendsSingleNameValue, err := GetPointerValue(friendsSingleNamePointer, originalJson)
		if err != nil {
			t.Fatalf("Failed to get pointer value: %v", err)
		}

		assert.Equal(t, friendsSingleNameValue, "Bob")
	})

	t.Run("nested multi array pointer", func(t *testing.T) {
		contactsMultiValuePointer, err := MakePointer("friends[*].contacts[*].value")
		if err != nil {
			t.Fatalf("Failed to create pointer: %v", err)
		}

		contactsMultiValues, err := GetPointerArrayValues(contactsMultiValuePointer, originalJson)
		if err != nil {
			t.Fatalf("Failed to get pointer value: %v", err)
		}

		assert.Equal(t, contactsMultiValues, []any{"alice@example.com", "123-456-7890", "987-654-3210", "bob@example.com"})
	})

	t.Run("string array collection", func(t *testing.T) {
		rolesPointer, err := MakePointer("roles[*]")
		if err != nil {
			t.Fatalf("Failed to create pointer: %v", err)
		}

		rolesArrayValues, err := GetPointerArrayValues(rolesPointer, originalJson)
		if err != nil {
			t.Fatalf("Failed to get pointer value: %v", err)
		}

		assert.Equal(t, rolesArrayValues, []any{"admin", "user", "editor", "guest", "moderator"})
	})
}
