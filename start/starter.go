package main

import (
	"context"
	"database/sql"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"log"
	"sync"
	"temporal-ledger-poc/app"
	appWorkflow "temporal-ledger-poc/app/appworkflow"
	"temporal-ledger-poc/app/db"
	"time"
)

func main() {
	c, err := client.Dial(client.Options{
		ConnectionOptions: client.ConnectionOptions{
			GetSystemInfoTimeout: 30 * time.Second,
		},
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()
	accounts := getAccounts()
	var wg sync.WaitGroup
	wg.Add(len(accounts))

	for i := 0; i < len(accounts); i++ {
		input := app.MoneyTransferWorkflowDetails{
			SourceAccount:   "schet_klienta",
			TargetAccount:   accounts[i],
			Amount:          100,
			TransactionType: "schet_klienta" + "_to_" + accounts[i],
		}
		go executeStarters(c, input, &wg)
	}

	wg.Wait()
	log.Println("All workflow started successfully!")
}

func executeStarters(c client.Client, input app.MoneyTransferWorkflowDetails, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()

	options := client.StartWorkflowOptions{
		ID:                       time.Now().Format("2024-01-02 15:04:05.000000"),
		TaskQueue:                app.MoneyTransferTaskQueueName,
		WorkflowIDConflictPolicy: enums.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	we, err := c.ExecuteWorkflow(ctx, options, appWorkflow.MoneyTransfer, input)
	if err != nil {
		log.Printf("Unable to start the Workflow [WorkflowID: %s]: %v", options.ID, err)
		return
	}

	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Printf("Unable to get Workflow result [WorkflowID: %s]: %v", we.GetID(), err)
		return
	}

	log.Printf("Workflow result [WorkflowID: %s]: %s", we.GetID(), result)
}

func getAccounts() []string {
	query := "SELECT account_number FROM accounts WHERE account_number <> 'schet_klienta'"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		log.Fatalf("Unable to execute query: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Unable to close rows: %v", err)
		}
	}(rows)

	accounts := make([]string, 0)

	for rows.Next() {
		var accountNumber string
		if err := rows.Scan(&accountNumber); err != nil {
			log.Fatalf("Unable to scan row: %v", err)
		}
		accounts = append(accounts, accountNumber)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Error during row iteration: %v", err)
	}

	return accounts
}
