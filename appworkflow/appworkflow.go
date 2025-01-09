package appworkflow

import (
	"fmt"
	"go.uber.org/multierr"
	"temporal-ledger-poc/app"
	"temporal-ledger-poc/app/activity"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MoneyTransfer(ctx workflow.Context, input app.MoneyTransferWorkflowDetails) (string, error) {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumAttempts:    3,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var transactionID string
	err := workflow.ExecuteActivity(ctx, activity.CreateTransaction, input).Get(ctx, &transactionID)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	var debitOperationId string
	err = workflow.ExecuteActivity(ctx, activity.Debit, input, transactionID).Get(ctx, &debitOperationId)
	if err != nil {
		return "", fmt.Errorf("debit failed: %w", err)
	}

	var creditOperationId string
	err = workflow.ExecuteActivity(ctx, activity.Credit, input, transactionID).Get(ctx, &creditOperationId)
	if err != nil {
		return "", fmt.Errorf("credit failed: %w", err)
	}
	defer func() {
		if err != nil {
			errCompensation := workflow.ExecuteActivity(ctx, activity.RollbackCredit, input, transactionID).Get(ctx, nil)
			err = multierr.Append(err, errCompensation)
		}
	}()

	return fmt.Sprintf("Transaction completed successfully. Transaction ID: %s, Debit Operation ID: %s,"+
		" Credit Operation ID: %s", transactionID, debitOperationId, creditOperationId), nil
}
