package activity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"temporal-ledger-poc/app"
	"temporal-ledger-poc/app/db"
)

func CreateTransaction(ctx context.Context, data app.MoneyTransferWorkflowDetails) (string, error) {
	log := activity.GetLogger(ctx)
	tx, err := db.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}

	var transactionId string
	queryCheck := `
		SELECT transaction_id 
		FROM transactions 
		WHERE transaction_type = $1
	`
	err = tx.QueryRowContext(ctx, queryCheck, data.TransactionType).Scan(&transactionId)
	if err == nil {
		log.Info(fmt.Sprintf("Transaction already exists: ID=%s, Type=%s\n", transactionId, data.TransactionType))
		return transactionId, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error checking existing transaction: %w", err)
	}

	transactionId = uuid.New().String()
	queryCreate := `
		INSERT INTO transactions (transaction_type, transaction_id, status, created_at)
		VALUES ($1, $2, 'NEW', NOW())
	`
	_, err = tx.Exec(queryCreate, data.TransactionType, transactionId)
	if err != nil {
		return "", fmt.Errorf("error creating transaction: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info(fmt.Sprintf("Transaction created: ID=%s, Type=%s\n", transactionId, data.TransactionType))
	return transactionId, nil
}

func Debit(ctx context.Context, data app.MoneyTransferWorkflowDetails, transactionId string) (string, error) {
	log := activity.GetLogger(ctx)
	tx, err := db.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}

	var debitOperationId string
	query := `SELECT operation_id FROM operations WHERE transaction_id = $1 AND operation_type = $2 AND account_number = $3 AND status = 'SUCCESS'`
	err = tx.QueryRowContext(ctx, query, transactionId, "debit", data.SourceAccount).Scan(&debitOperationId)
	if err == nil {
		log.Info(fmt.Sprintf("Debit operation already exists for transaction %s", transactionId))
		return debitOperationId, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		err = errors.Unwrap(err)
		return "", fmt.Errorf("failed to check debit operation existence: %w", err)
	}

	debitOperationId = uuid.New().String()
	query = `
		INSERT INTO operations (operation_id, transaction_id, account_number, operation_type, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'SUCCESS', NOW())
		RETURNING id`
	_, err = tx.Exec(query, debitOperationId, transactionId, data.SourceAccount, "debit", data.Amount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("failed to create debit operation: %w", err)
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE account_number = $2", data.Amount, data.SourceAccount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	_, err = tx.Exec("UPDATE transactions SET status = 'PENDING' WHERE transaction_id = $1", transactionId)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return debitOperationId, nil
}

func Credit(ctx context.Context, data app.MoneyTransferWorkflowDetails, transactionId string) (string, error) {
	log := activity.GetLogger(ctx)
	tx, err := db.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}

	var creditOperationId string
	query := `SELECT operation_id FROM operations WHERE transaction_id = $1 AND operation_type = $2 AND account_number = $3 AND status = 'SUCCESS'`
	err = tx.QueryRowContext(ctx, query, transactionId, "credit", data.TargetAccount).Scan(&creditOperationId)
	if err == nil {
		log.Info(fmt.Sprintf("Credit operation already exists for transaction %s", transactionId))
		return creditOperationId, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("failed to check credit operation existence: %w", err)
	}

	creditOperationId = uuid.New().String()
	query = `
		INSERT INTO operations (operation_id, transaction_id, account_number, operation_type, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'SUCCESS', NOW())
		RETURNING id`
	_, err = tx.Exec(query, creditOperationId, transactionId, data.TargetAccount, "credit", data.Amount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("failed to create credit operation: %w", err)
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE account_number = $2", data.Amount, data.TargetAccount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	_, err = tx.Exec("UPDATE transactions SET status = 'SUCCESS' WHERE transaction_id = $1", transactionId)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return creditOperationId, nil
}

func RollbackCredit(ctx context.Context, data app.MoneyTransferWorkflowDetails, transactionId string) (string, error) {
	log := activity.GetLogger(ctx)
	tx, err := db.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}

	var refundOperationId string
	query := `SELECT operation_id FROM operations WHERE transaction_id = $1 AND operation_type = $2 AND status = 'SUCCESS'`
	err = tx.QueryRowContext(ctx, query, transactionId, "rollback").Scan(&refundOperationId)
	if err == nil {
		log.Info(fmt.Sprintf("RollbackCredit already exists for transaction %s", transactionId))
		return refundOperationId, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("failed to check operation existence: %w", err)
	}

	refundOperationId = uuid.New().String()
	query = `
		INSERT INTO operations (operation_id, transaction_id, account_number, operation_type, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'SUCCESS', NOW())
		RETURNING id`
	_, err = tx.Exec(query, refundOperationId, transactionId, data.SourceAccount, "rollback", data.Amount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("failed to create refund operation: %w", err)
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE account_number = $2", data.Amount, data.SourceAccount)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	_, err = tx.Exec("UPDATE transactions SET status = 'ROLLBACK_SUCCESS' WHERE transaction_id = $1", transactionId)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit refund transaction: %w", err)
	}

	return refundOperationId, nil
}
