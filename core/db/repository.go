package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository provides data access for a model type
type Repository[T Modeler] struct {
	client    *Client
	tableName string
}

// NewRepository creates a new repository for a model type
func NewRepository[T Modeler](client *Client) *Repository[T] {
	var model T
	return &Repository[T]{
		client:    client,
		tableName: model.TableName(),
	}
}

// Client returns the underlying database client
func (r *Repository[T]) Client() *Client {
	return r.client
}

// TableName returns the table name for this repository
func (r *Repository[T]) TableName() string {
	return r.tableName
}

// New creates a new record ready for saving
func (r *Repository[T]) New() *Record[T] {
	var model T
	// Create a pointer to a new instance
	modelPtr := reflect.New(reflect.TypeOf(model).Elem()).Interface().(T)
	return &Record[T]{
		model: modelPtr,
		repo:  r,
	}
}

// Wrap wraps an existing model in a Record
func (r *Repository[T]) Wrap(model T) *Record[T] {
	return &Record[T]{
		model: model,
		repo:  r,
	}
}

// Create inserts a new record
func (r *Repository[T]) Create(ctx context.Context, model T) error {
	// Run before create hooks
	if err := RunBeforeCreateHooks(model); err != nil {
		return err
	}

	// Apply base model defaults if it embeds Model
	if baseModel := getBaseModel(model); baseModel != nil {
		ApplyBeforeCreate(baseModel)
	}

	// Build and execute insert query
	query, _ := r.buildInsertQuery(model)
	_, err := r.client.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	// Run after create hooks
	if err := RunAfterCreateHooks(model); err != nil {
		return err
	}

	return nil
}

// Update saves changes to an existing record
func (r *Repository[T]) Update(ctx context.Context, model T) error {
	// Check if persisted
	if baseModel := getBaseModel(model); baseModel != nil {
		if baseModel.IsNew() {
			return ErrNotPersisted
		}
	}

	// Run before update hooks
	if err := RunBeforeUpdateHooks(model); err != nil {
		return err
	}

	// Apply base model updates
	if baseModel := getBaseModel(model); baseModel != nil {
		ApplyBeforeUpdate(baseModel)
	}

	// Build and execute update query
	query, _ := r.buildUpdateQuery(model)
	result, err := r.client.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	// Run after update hooks
	if err := RunAfterUpdateHooks(model); err != nil {
		return err
	}

	return nil
}

// Save creates or updates the record based on whether it's new
func (r *Repository[T]) Save(ctx context.Context, model T) error {
	if baseModel := getBaseModel(model); baseModel != nil {
		if baseModel.IsNew() {
			return r.Create(ctx, model)
		}
	}
	return r.Update(ctx, model)
}

// Delete soft-deletes a record
func (r *Repository[T]) Delete(ctx context.Context, model T) error {
	id := getModelID(model)
	if id == "" {
		return ErrNotPersisted
	}

	// Run before delete hooks
	if err := RunBeforeDeleteHooks(model); err != nil {
		return err
	}

	// Soft delete
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", r.tableName)
	query = r.client.Rebind(query)

	result, err := r.client.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	// Update model's DeletedAt
	if baseModel := getBaseModel(model); baseModel != nil {
		ApplyBeforeDelete(baseModel)
	}

	// Run after delete hooks
	if err := RunAfterDeleteHooks(model); err != nil {
		return err
	}

	return nil
}

// HardDelete permanently deletes a record
func (r *Repository[T]) HardDelete(ctx context.Context, model T) error {
	id := getModelID(model)
	if id == "" {
		return ErrNotPersisted
	}

	// Run before delete hooks
	if err := RunBeforeDeleteHooks(model); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.tableName)
	query = r.client.Rebind(query)

	result, err := r.client.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("hard delete failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	// Run after delete hooks
	if err := RunAfterDeleteHooks(model); err != nil {
		return err
	}

	return nil
}

// Restore un-deletes a soft-deleted record
func (r *Repository[T]) Restore(ctx context.Context, model T) error {
	id := getModelID(model)
	if id == "" {
		return ErrNotPersisted
	}

	query := fmt.Sprintf("UPDATE %s SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL", r.tableName)
	query = r.client.Rebind(query)

	result, err := r.client.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	// Update model
	if baseModel := getBaseModel(model); baseModel != nil {
		baseModel.Restore()
		ApplyBeforeUpdate(baseModel)
	}

	return nil
}

// FindByID retrieves a record by primary key
func (r *Repository[T]) FindByID(ctx context.Context, id string) (*Record[T], error) {
	var model T
	// Create a new instance to scan into
	modelPtr := reflect.New(reflect.TypeOf(model).Elem()).Interface()

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", r.tableName)
	query = r.client.Rebind(query)

	err := r.client.db.GetContext(ctx, modelPtr, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find by id failed: %w", err)
	}

	typedModel := modelPtr.(T)

	// Run after find hooks
	if err := RunAfterFindHooks(typedModel); err != nil {
		return nil, err
	}

	return &Record[T]{model: typedModel, repo: r}, nil
}

// FindAll retrieves records with optional query options
func (r *Repository[T]) FindAll(ctx context.Context, opts ...QueryOption) ([]*Record[T], error) {
	qb := NewQueryBuilder(r.tableName)
	qb.Apply(opts...)

	query, args := qb.Build()
	query = r.client.Rebind(query)

	// Create a slice to scan into
	var model T
	sliceType := reflect.SliceOf(reflect.TypeOf(model).Elem())
	modelsPtr := reflect.New(sliceType)

	err := r.client.db.SelectContext(ctx, modelsPtr.Interface(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("find all failed: %w", err)
	}

	models := modelsPtr.Elem()
	records := make([]*Record[T], models.Len())
	for i := 0; i < models.Len(); i++ {
		m := models.Index(i).Addr().Interface().(T)
		// Run after find hooks
		if err := RunAfterFindHooks(m); err != nil {
			return nil, err
		}
		records[i] = &Record[T]{model: m, repo: r}
	}

	return records, nil
}

// FindOne retrieves the first matching record
func (r *Repository[T]) FindOne(ctx context.Context, opts ...QueryOption) (*Record[T], error) {
	opts = append(opts, Limit(1))
	records, err := r.FindAll(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, ErrNotFound
	}
	return records[0], nil
}

// First returns the first record ordered by created_at
func (r *Repository[T]) First(ctx context.Context, opts ...QueryOption) (*Record[T], error) {
	opts = append(opts, OrderByAsc("created_at"), Limit(1))
	return r.FindOne(ctx, opts...)
}

// Last returns the last record ordered by created_at
func (r *Repository[T]) Last(ctx context.Context, opts ...QueryOption) (*Record[T], error) {
	opts = append(opts, OrderByDesc("created_at"), Limit(1))
	return r.FindOne(ctx, opts...)
}

// Count returns the number of matching records
func (r *Repository[T]) Count(ctx context.Context, opts ...QueryOption) (int64, error) {
	qb := NewQueryBuilder(r.tableName)
	qb.Apply(opts...)

	query, args := qb.BuildCount()
	query = r.client.Rebind(query)

	var count int64
	err := r.client.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// Exists checks if a record with the given ID exists
func (r *Repository[T]) Exists(ctx context.Context, id string) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ? AND deleted_at IS NULL)", r.tableName)
	query = r.client.Rebind(query)

	var exists bool
	err := r.client.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("exists check failed: %w", err)
	}

	return exists, nil
}

// ExistsWhere checks if any record matches the conditions
func (r *Repository[T]) ExistsWhere(ctx context.Context, opts ...QueryOption) (bool, error) {
	count, err := r.Count(ctx, opts...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteWhere soft-deletes all records matching the conditions
func (r *Repository[T]) DeleteWhere(ctx context.Context, opts ...QueryOption) (int64, error) {
	qb := NewQueryBuilder(r.tableName)
	qb.Apply(opts...)

	// Build WHERE clause
	conditions := []string{"deleted_at IS NULL"}
	var args []any
	args = append(args, time.Now())

	for _, cond := range qb.conditions {
		conditions = append(conditions, cond)
	}
	args = append(args, qb.args...)

	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE %s",
		r.tableName, strings.Join(conditions, " AND "))
	query = r.client.Rebind(query)

	result, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete where failed: %w", err)
	}

	return result.RowsAffected()
}

// UpdateWhere updates all records matching the conditions
func (r *Repository[T]) UpdateWhere(ctx context.Context, updates map[string]any, opts ...QueryOption) (int64, error) {
	if len(updates) == 0 {
		return 0, nil
	}

	qb := NewQueryBuilder(r.tableName)
	qb.Apply(opts...)

	// Build SET clause
	setParts := make([]string, 0, len(updates)+1)
	args := make([]any, 0, len(updates)+len(qb.args)+1)

	for col, val := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	// Always update updated_at
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())

	// Build WHERE clause
	conditions := []string{"deleted_at IS NULL"}
	conditions = append(conditions, qb.conditions...)
	args = append(args, qb.args...)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		r.tableName,
		strings.Join(setParts, ", "),
		strings.Join(conditions, " AND "))
	query = r.client.Rebind(query)

	result, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("update where failed: %w", err)
	}

	return result.RowsAffected()
}

// WithTx returns a new repository using the given transaction
func (r *Repository[T]) WithTx(tx *sqlx.Tx) *TxRepository[T] {
	return &TxRepository[T]{
		repo: r,
		tx:   tx,
	}
}

// Transaction executes a function within a transaction
func (r *Repository[T]) Transaction(ctx context.Context, fn func(*TxRepository[T]) error) error {
	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	txRepo := r.WithTx(tx)

	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// Helper methods

func (r *Repository[T]) buildInsertQuery(model T) (string, []any) {
	columns, placeholders := getInsertColumns(model)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		r.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))
	return query, nil
}

func (r *Repository[T]) buildUpdateQuery(model T) (string, []any) {
	columns := getUpdateColumns(model)
	setParts := make([]string, len(columns))
	for i, col := range columns {
		setParts[i] = fmt.Sprintf("%s = :%s", col, col)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = :id AND deleted_at IS NULL",
		r.tableName,
		strings.Join(setParts, ", "))
	return query, nil
}

// getBaseModel extracts the embedded Model from a struct if present
func getBaseModel(model any) *Model {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	// Look for embedded Model field
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == reflect.TypeOf(Model{}) {
			if field.CanAddr() {
				return field.Addr().Interface().(*Model)
			}
		}
	}

	return nil
}

// getModelID extracts the ID from a model
func getModelID(model any) string {
	if getter, ok := model.(IDGetter); ok {
		return getter.GetID()
	}
	if baseModel := getBaseModel(model); baseModel != nil {
		return baseModel.ID
	}
	return ""
}

// getInsertColumns returns columns and placeholders for INSERT
func getInsertColumns(model any) ([]string, []string) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var columns, placeholders []string

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs
		if field.Anonymous {
			embeddedV := v.Field(i)
			if embeddedV.Kind() == reflect.Struct {
				embeddedCols, embeddedPlaceholders := getInsertColumns(embeddedV.Interface())
				columns = append(columns, embeddedCols...)
				placeholders = append(placeholders, embeddedPlaceholders...)
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, ":"+dbTag)
	}

	return columns, placeholders
}

// getUpdateColumns returns columns for UPDATE (excludes id, created_at)
func getUpdateColumns(model any) []string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var columns []string

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs
		if field.Anonymous {
			embeddedV := v.Field(i)
			if embeddedV.Kind() == reflect.Struct {
				embeddedCols := getUpdateColumns(embeddedV.Interface())
				columns = append(columns, embeddedCols...)
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" || dbTag == "id" || dbTag == "created_at" {
			continue
		}

		columns = append(columns, dbTag)
	}

	return columns
}

// TxRepository is a repository bound to a transaction
type TxRepository[T Modeler] struct {
	repo *Repository[T]
	tx   *sqlx.Tx
}

// Create inserts a new record within the transaction
func (r *TxRepository[T]) Create(ctx context.Context, model T) error {
	if err := RunBeforeCreateHooks(model); err != nil {
		return err
	}

	if baseModel := getBaseModel(model); baseModel != nil {
		ApplyBeforeCreate(baseModel)
	}

	query, _ := r.repo.buildInsertQuery(model)
	_, err := r.tx.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	return RunAfterCreateHooks(model)
}

// Update saves changes within the transaction
func (r *TxRepository[T]) Update(ctx context.Context, model T) error {
	if err := RunBeforeUpdateHooks(model); err != nil {
		return err
	}

	if baseModel := getBaseModel(model); baseModel != nil {
		ApplyBeforeUpdate(baseModel)
	}

	query, _ := r.repo.buildUpdateQuery(model)
	result, err := r.tx.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return RunAfterUpdateHooks(model)
}

// FindByID retrieves a record within the transaction
func (r *TxRepository[T]) FindByID(ctx context.Context, id string) (*Record[T], error) {
	var model T
	modelPtr := reflect.New(reflect.TypeOf(model).Elem()).Interface()

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", r.repo.tableName)
	query = r.repo.client.Rebind(query)

	err := r.tx.GetContext(ctx, modelPtr, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find by id failed: %w", err)
	}

	typedModel := modelPtr.(T)

	if err := RunAfterFindHooks(typedModel); err != nil {
		return nil, err
	}

	return &Record[T]{model: typedModel, repo: r.repo}, nil
}
