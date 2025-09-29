# Lazy Lagoon Service

A Go-based service for processing CSV, JSON, JSONL, and SQL files with pagination and transformation capabilities.

## Overview

The Lazy Lagoon provides two main endpoints:

- **Paginate**: Splits large files (CSV, JSONL, SQL) into smaller chunks for preview purposes
- **Transform**: Applies transformation rules to data files using expression-based filtering

## API Endpoints

### 1. Paginate Endpoint

**URL**: `POST /paginate`

Splits large data files into smaller chunks for preview and pagination purposes.

#### Supported Data Types

- **CSV**: Comma-separated values files
- **JSONL**: JSON Lines format (one JSON object per line)
- **SQL**: SQL query result files
- **JSON**: Limited support (returns attributes only, no pagination)

#### Request Body Structure

```json
{
  "input": {
    "storageType": "s3",
    "dataType": "CSV",
    "reference": {
      "id": "input-file-id",
      "bucket": "my-input-bucket", 
      "prefix": "path/to/data.csv",
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret-key",
        "accessToken": "aws-access-token"
      },
      "resources": {
        "id": "resource-id"
      }
    }
  },
  "output": {
    "storageType": "s3",
    "dataType": "CSV", 
    "reference": {
      "id": "output-file-id",
      "bucket": "my-output-bucket",
      "prefix": "output/paginated",
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret-key",
        "accessToken": "aws-access-token"
      },
      "resources": {
        "id": "resource-id"
      }
    }
  }
}
```

#### Expected Response

```json
{
  "message": "Success: 15 chunks paginated",
  "totalPages": 15,
  "attributes": {
    "paths": [
      "name",
      "email", 
      "age",
      "address.street",
      "address.city"
    ]
  }
}
```

#### Output Files Created

For a CSV file with 750 rows (50 rows per chunk), the following files will be created:

- `output/`key `/pages/1.csv` - Contains header + rows 1-49
- `output/`key `/pages/2.csv` - Contains header + rows 50-99
- `output/`key `/pages/3.csv` - Contains header + rows 100-149
- ...and so on

For JSONL files:

- `output/key/pages/1.jsonl` - Contains first 50 JSON objects
- `output/`key `/pages/2.jsonl` - Contains next 50 JSON objects
- ...and so on

### 2. Transform Endpoint

**URL**: `POST /transform`

Applies transformation rules to data files using expression-based filtering and actions.

#### Supported Data Types

- **CSV**: Comma-separated values files
- **SQL**: SQL query result files
- **JSON**: JSON format files
- **JSONL**: JSON Lines format files

#### Request Body Structure

```json
{
  "input": {
    "storageType": "s3",
    "dataType": "CSV",
    "reference": {
      "id": "input-file-id",
      "bucket": "my-input-bucket",
      "prefix": "path/to/employees.csv", 
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret-key",
        "accessToken": "aws-access-token"
      },
      "resources": {
        "id": "resource-id"
      }
    }
  },
  "output": {
    "storageType": "s3",
    "dataType": "CSV",
    "reference": {
      "id": "output-file-id", 
      "bucket": "my-output-bucket",
      "prefix": "path/to/employees_transformed.csv",
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret-key",
        "accessToken": "aws-access-token"
      },
      "resources": {
        "id": "resource-id"
      }
    }
  },
  "rules": [
    {
      "expression": {
        "logicalOperator": "AND",
        "expressions": [
          {
            "fieldName": "department",
            "operator": "equals",
            "value": "HR"
          },
          {
            "fieldName": "salary",
            "operator": "greaterThan", 
            "value": 50000
          }
        ]
      },
      "actions": [
        {
          "actionType": "redact",
          "fieldName": "ssn"
        },
        {
          "actionType": "redact", 
          "fieldName": "salary"
        }
      ]
    }
  ],
  "webhook": {
    "url": "https://my-app.com/webhook",
    "responseToken": "webhook-token-123",
    "payload": {
      "msg": "Transform completed successfully",
      "browserTabID": "tab-456",
      "uuid": "transform-789",
      "userId": "user-123",
      "s3Bucket": "my-output-bucket", 
      "s3Key": "path/to/employees_transformed.csv",
      "sourceId": "source-123",
      "status": "completed"
    }
  }
}
```

#### Expression Operators

- **`"equals"`**: Exact match (==)
- **`"notEquals"`**: Not equal (!=)
- **`"greaterThan"`**: Greater than (>)
- **`"greaterThanOrEqual"`**: Greater than or equal (>=)
- **`"lessThan"`**: Less than (<)
- **`"lessThanOrEqual"`**: Less than or equal (<=)
- **`"contains"`**: String contains substring
- **`"startsWith"`**: String starts with value
- **`"endsWith"`**: String ends with value

#### Logical Operators

- **`"AND"`**: All expressions must be true
- **`"OR"`**: At least one expression must be true

#### Action Types

- **`"redact"`**: Replace field value with `[REDACTED]`
- **`"exclude"`**: Remove the field entirely from output

#### Expected Response

```json
"Completed transformation"
```

## Example Use Cases

### Example 1: Paginating a Large CSV File

**Request:**

```json
{
  "input": {
    "storageType": "s3",
    "dataType": "CSV", 
    "reference": {
      "bucket": "data-lake",
      "prefix": "sales/2024/transactions.csv",
      "region": "us-west-2"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret"
      }
    }
  },
  "output": {
    "storageType": "s3",
    "dataType": "CSV",
    "reference": {
      "bucket": "preview-data",
      "prefix": "sales-preview",
      "region": "us-west-2" 
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret"
      }
    }
  }
}
```

**Response:**

```json
{
  "message": "Success: 25 chunks paginated",
  "totalPages": 25,
  "attributes": {
    "paths": [
      "transaction_id",
      "customer_id", 
      "amount",
      "date",
      "product_category"
    ]
  }
}
```

### Example 2: Redacting Sensitive Information

**Request:**

```json
{
  "input": {
    "storageType": "s3",
    "dataType": "JSONL",
    "reference": {
      "bucket": "user-data",
      "prefix": "profiles/users.jsonl",
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret"
      }
    }
  },
  "output": {
    "storageType": "s3", 
    "dataType": "JSONL",
    "reference": {
      "bucket": "sanitized-data",
      "prefix": "profiles/users_clean.jsonl",
      "region": "us-east-1"
    },
    "credential": {
      "secrets": {
        "secret": "aws-secret"
      }
    }
  },
  "rules": [
    {
      "expression": {
        "logicalOperator": "OR",
        "expressions": [
          {
            "fieldName": "age",
            "operator": "lessThan",
            "value": 18
          },
          {
            "fieldName": "consent",
            "operator": "equals", 
            "value": false
          }
        ]
      },
      "actions": [
        {
          "actionType": "redact",
          "fieldName": "email"
        },
        {
          "actionType": "redact",
          "fieldName": "phone"
        },
        {
          "actionType": "exclude",
          "fieldName": "ssn"
        }
      ]
    }
  ]
}
```

## Request Body Field Descriptions

### Storage Types

Currently supported: `"s3"`

### Data Types

- **`"CSV"`**: Comma-separated values
- **`"JSONL"`**: JSON Lines (one JSON object per line)
- **`"JSON"`**: Standard JSON format
- **`"SQL"`**: SQL query results

### Storage Reference Fields

#### S3 Storage

- **`bucket`**: S3 bucket name
- **`prefix`**: File path within the bucket
- **`region`**: AWS region (e.g., "us-east-1")

#### Credentials

- **`secrets.secret`**: AWS secret access key
- **`secrets.accessToken`**: AWS access token (if using temporary credentials)
- **`resources.id`**: Resource identifier

### Rule Structure

#### Expression

- **`logicalOperator`**: "AND" or "OR" for combining multiple expressions
- **`expressions`**: Array of individual filter conditions

#### Expression Fields

- **`fieldName`**: Name of the field/column to evaluate
- **`operator`**: Comparison operator (equals, greaterThan, contains, etc.)
- **`value`**: Value to compare against (string, number, boolean)

#### Actions

- **`actionType`**: Type of action to perform ("redact" or "exclude")
- **`fieldName`**: Target field for the action

### Webhook (Optional)

Optional callback configuration for async processing notifications.

- **`url`**: Webhook endpoint URL
- **`responseToken`**: Authentication token for webhook
- **`payload`**: Custom payload to send to webhook

## Pagination Details

### CSV Pagination

- Default chunk size: 50 rows per page
- Header row is included in every page
- Files are stored as `.csv` format

### JSONL Pagination

- Default chunk size: 50 JSON objects per page
- Each page maintains valid JSONL format
- Files are stored as `.jsonl` format

## Error Handling

The service returns appropriate HTTP status codes:

- **200**: Transform completed successfully
- **202**: Paginate completed successfully
- **400**: Bad Request (validation errors, unsupported data types)
- **500**: Internal Server Error (processing failures, storage errors)

Error responses include detailed error messages and may include rule/action/expression indices for transformation errors.

## Limitations

- Maximum chunk size: 50 rows/objects per page for pagination
- JSON files only return field paths for pagination (no actual chunking)
- Webhook notifications are optional and only sent after successful transforms
- Expression evaluation supports basic data types (string, number, boolean)
