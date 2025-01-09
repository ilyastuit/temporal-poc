package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
	"temporal-ledger-poc/app"
	"temporal-ledger-poc/app/activity"
	"temporal-ledger-poc/app/appworkflow"
	"time"
)

func main() {
	c, err := client.Dial(client.Options{
		ConnectionOptions: client.ConnectionOptions{
			GetSystemInfoTimeout: 30 * time.Second,
		},
	})
	if err != nil {
		log.Fatal("Unable to create Temporal client.", err)
	}
	defer func(c client.Client) {
		if c != nil {
			c.Close()
		}
	}(c)

	w := worker.New(c, app.MoneyTransferTaskQueueName, worker.Options{
		MaxConcurrentActivityTaskPollers:       25,
		MaxConcurrentWorkflowTaskPollers:       25,
		MaxConcurrentActivityExecutionSize:     50,
		MaxConcurrentWorkflowTaskExecutionSize: 50,
	})

	// This worker hosts both Workflow and Activity functions.
	w.RegisterWorkflow(appworkflow.MoneyTransfer)
	w.RegisterActivity(activity.CreateTransaction)
	w.RegisterActivity(activity.Debit)
	w.RegisterActivity(activity.Credit)
	w.RegisterActivity(activity.RollbackCredit)

	// Start listening to the Task Queue.
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatal("unable to start Worker", err)
	}
}
