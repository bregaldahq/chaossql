package evaluator

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

// Evaluator runs invariant queries and validates expressions.
type Evaluator struct {}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate executes the invariant's SQL query and evaluates the assert expression.
func (e *Evaluator) Evaluate(ctx context.Context, driver drivers.DatabaseDriver, inv domain.InvariantConfig) (domain.InvariantResult, error) {
	result := domain.InvariantResult{
		Name:         inv.Name,
		Expression:   inv.Assert,
		ActualValues: make(map[string]interface{}),
	}

	rows, err := driver.Query(ctx, inv.Query)
	if err != nil {
		result.Error = fmt.Errorf("invariant query failed: %w", err)
		return result, result.Error
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		result.Error = fmt.Errorf("failed to get column names: %w", err)
		return result, result.Error
	}

	if !rows.Next() {
		result.Error = fmt.Errorf("invariant query returned no rows")
		return result, result.Error
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		result.Error = fmt.Errorf("failed to scan invariant row: %w", err)
		return result, result.Error
	}

	env := make(map[string]interface{})
	for i, colName := range cols {
		val := values[i]
		// Convert bytes to string if needed
		if b, ok := val.([]byte); ok {
			val = string(b)
		}
		env[colName] = val
		result.ActualValues[colName] = val
	}

	// Compile and run expression safely
	program, err := expr.Compile(inv.Assert, expr.Env(env), expr.AsBool())
	if err != nil {
		result.Error = fmt.Errorf("invalid assert expression '%s': %w", inv.Assert, err)
		return result, result.Error
	}

	output, err := expr.Run(program, env)
	if err != nil {
		result.Error = fmt.Errorf("expression evaluation failed: %w", err)
		return result, result.Error
	}

	boolVal, ok := output.(bool)
	if !ok {
		result.Error = fmt.Errorf("expression did not return a boolean")
		return result, result.Error
	}

	result.Passed = boolVal
	return result, nil
}
