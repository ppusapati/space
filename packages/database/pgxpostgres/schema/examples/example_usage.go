package schema_example

// This file provides example usage patterns for schema-based column validation.
// DO NOT import this file in production code - it's for documentation purposes only.

import (
	"context"
	"log"

	"p9e.in/samavaya/packages/database/pgxpostgres/operations"
	"p9e.in/samavaya/packages/database/pgxpostgres/schema"
	"p9e.in/samavaya/packages/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Example 1: Basic Schema Registration
func ExampleBasicRegistration() {
	// Define table schema
	userSchema := schema.NewTableSchema("users", []string{
		"id",
		"uuid",
		"name",
		"email",
		"password_hash",
		"created_at",
		"updated_at",
		"deleted_at",
		"is_active",
		"created_by",
		"updated_by",
		"deleted_by",
	})

	// Register globally (do this once at application startup)
	schema.RegisterSchema("users", userSchema)

	log.Println("User schema registered successfully")
}

// Example 2: Using Aliases for Backward Compatibility
func ExampleAliasRegistration() {
	// Create schema
	userSchema := schema.NewTableSchema("users", []string{
		"id", "name", "email",
	})

	// Register aliases for legacy field names
	userSchema.RegisterAlias("user_id", "id")
	userSchema.RegisterAlias("user_name", "name")
	userSchema.RegisterAlias("user_email", "email")

	schema.RegisterSchema("users", userSchema)

	// Now both old and new field names work
	log.Println("Schema registered with aliases")
}

// Example 3: Valid Query - Passes Validation
func ExampleValidQuery(ctx context.Context, pool *pgxpool.Pool) {
	// Register schema first
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "name", "email", "created_at",
	}))

	// This query will pass validation
	dm := models.DataModel[User]{
		TableName:  "users",
		FieldNames: []string{"id", "name", "email"},
		Where:      "is_active = $1",
		WhereArgs:  []any{true},
	}

	result, err := operations.ExecuteQuery(ctx, pool, &dm, operations.QueryTypeSelect)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	log.Printf("Query succeeded: %+v", result)
}

// Example 4: Invalid Query - Fails Validation
func ExampleInvalidQuery(ctx context.Context, pool *pgxpool.Pool) {
	// Register schema
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "name", "email", "created_at",
	}))

	// This query has typos and will fail validation
	dm := models.DataModel[User]{
		TableName:  "users",
		FieldNames: []string{"id", "naem", "emial"}, // Typos!
	}

	result, err := operations.ExecuteQuery(ctx, pool, &dm, operations.QueryTypeSelect)
	if err != nil {
		// Error message will be:
		// "invalid column(s) 'naem, emial' for table 'users'.
		//  Valid columns: created_at, email, id, name"
		log.Printf("Validation failed (expected): %v", err)
		return
	}

	log.Printf("Query succeeded: %+v", result)
}

// Example 5: Using Wildcard Selection
func ExampleWildcardQuery(ctx context.Context, pool *pgxpool.Pool) {
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "name", "email",
	}))

	// Wildcard is always valid
	dm := models.DataModel[User]{
		TableName:  "users",
		FieldNames: []string{"*"},
		Where:      "id = $1",
		WhereArgs:  []any{123},
	}

	result, err := operations.ExecuteQuery(ctx, pool, &dm, operations.QueryTypeSelect)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	log.Printf("Query succeeded: %+v", result)
}

// Example 6: Dynamic Schema Updates
func ExampleDynamicSchemaUpdate() {
	// Initial schema
	userSchema := schema.NewTableSchema("users", []string{
		"id", "name", "email",
	})
	schema.RegisterSchema("users", userSchema)

	// Later, add computed fields
	userSchema.AddColumn("full_name")
	userSchema.AddColumn("avatar_url")

	// Remove deprecated fields
	userSchema.RemoveColumn("old_field")

	log.Println("Schema updated dynamically")
}

// Example 7: Checking Schema Registration
func ExampleCheckSchemaRegistration() {
	// List all registered tables
	tables := schema.ListRegisteredTables()
	log.Printf("Registered tables: %v", tables)

	// Get specific schema
	userSchema := schema.GetSchema("users")
	if userSchema == nil {
		log.Println("User schema not registered")
		return
	}

	// Get valid columns
	columns := userSchema.GetValidColumnsList()
	log.Printf("Valid columns for users: %v", columns)
}

// Example 8: Manual Validation (Before Query Execution)
func ExampleManualValidation() {
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "name", "email",
	}))

	// Validate field list manually
	fieldsToQuery := []string{"id", "name", "invalid"}
	err := schema.ValidateTableColumns("users", fieldsToQuery)
	if err != nil {
		log.Printf("Validation failed: %v", err)
		// Handle error: show user valid columns, etc.
		return
	}

	log.Println("Fields are valid, proceed with query")
}

// Example 9: Graceful Degradation (Schema Not Registered)
func ExampleGracefulDegradation(ctx context.Context, pool *pgxpool.Pool) {
	// Schema NOT registered
	schema.ClearRegistry()

	// Query still works, but no schema validation
	dm := models.DataModel[User]{
		TableName:  "users",
		FieldNames: []string{"id", "invalid_field"}, // Won't be caught!
		Where:      "id = $1",
		WhereArgs:  []any{123},
	}

	// Validation passes (no schema registered)
	// Error will occur at database execution time
	_, err := operations.ExecuteQuery(ctx, pool, &dm, operations.QueryTypeSelect)
	if err != nil {
		log.Printf("Database error (not caught by validation): %v", err)
	}
}

// Example 10: Complete Application Setup
func ExampleApplicationSetup() {
	// In your main.go or init() function

	// Register all table schemas
	registerUserSchema()
	registerPostSchema()
	registerCommentSchema()
	registerCategorySchema()

	log.Println("All schemas registered successfully")
}

func registerUserSchema() {
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "uuid", "name", "email", "password_hash",
		"created_at", "updated_at", "deleted_at",
		"is_active", "created_by", "updated_by", "deleted_by",
	}))
}

func registerPostSchema() {
	postSchema := schema.NewTableSchema("posts", []string{
		"id", "uuid", "user_id", "title", "content", "slug",
		"created_at", "updated_at", "deleted_at",
		"is_active", "created_by", "updated_by", "deleted_by",
	})

	// Add aliases
	postSchema.RegisterAlias("author_id", "user_id")

	schema.RegisterSchema("posts", postSchema)
}

func registerCommentSchema() {
	schema.RegisterSchema("comments", schema.NewTableSchema("comments", []string{
		"id", "uuid", "post_id", "user_id", "content",
		"created_at", "updated_at", "deleted_at",
		"is_active", "created_by", "updated_by", "deleted_by",
	}))
}

func registerCategorySchema() {
	schema.RegisterSchema("categories", schema.NewTableSchema("categories", []string{
		"id", "uuid", "name", "slug", "description",
		"created_at", "updated_at", "deleted_at",
		"is_active", "created_by", "updated_by", "deleted_by",
	}))
}

// Example User struct
type User struct {
	ID           int64  `db:"id"`
	UUID         string `db:"uuid"`
	Name         string `db:"name"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	IsActive     bool   `db:"is_active"`
}

// Example 11: Integration with Field Mask from gRPC
func ExampleFieldMaskIntegration(ctx context.Context, pool *pgxpool.Pool, fieldMaskPaths []string) {
	schema.RegisterSchema("users", schema.NewTableSchema("users", []string{
		"id", "name", "email", "created_at",
	}))

	// Validate field mask from gRPC request
	err := schema.ValidateTableColumns("users", fieldMaskPaths)
	if err != nil {
		log.Printf("Invalid field mask from client: %v", err)
		// Return gRPC error: codes.InvalidArgument
		return
	}

	// Build query with validated fields
	dm := models.DataModel[User]{
		TableName:  "users",
		FieldNames: fieldMaskPaths,
	}

	result, err := operations.ExecuteQuerySlice(ctx, pool, &dm, operations.QueryTypeSelect)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	log.Printf("Query succeeded with %d results", len(result))
}

// Example 12: Testing with Schema Validation
func ExampleTestingWithSchema() {
	// In your test setup
	schema.ClearRegistry() // Clean state

	// Register test schema
	testSchema := schema.NewTableSchema("test_users", []string{
		"id", "name", "email",
	})
	schema.RegisterSchema("test_users", testSchema)

	// Run tests...
	log.Println("Test schema registered")

	// Cleanup after test
	schema.ClearRegistry()
}
