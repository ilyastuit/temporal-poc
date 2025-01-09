# Temporal Go Project Template

This is a simple project for demonstrating Temporal with the Go SDK.

The full 10-minute tutorial is here: https://learn.temporal.io/getting_started/go/first_program_in_go/

## Basic instructions

### Step 0: Temporal Server

Make sure [Temporal Server is running](https://docs.temporal.io/docs/server/quick-install/) first:

```bash
cd  docker
docker-compose up
```

### Step 1: Run migrations

```bash
sql-migrate up -limit=0 -config=dbconfig.yml
```

### Step 2: Run the Workflow

```bash
go run start/main.go
```

Observe that Temporal Web reflects the workflow, but it is still in "Running" status. This is because there is no Workflow or Activity Worker yet listening to the `TRANSFER_MONEY_TASK_QUEUE` task queue to process this work.

### Step 3: Run the Worker

In YET ANOTHER terminal instance, run the worker. Notice that this worker hosts both Workflow and Activity functions.

```bash
go run worker/main.go
```

Now you can see the workflow run to completion. You can also see the worker polling for workflows and activities in the task queue at [http://localhost:8080/namespaces/default/task-queues/TRANSFER_MONEY_TASK_QUEUE](http://localhost:8080/namespaces/default/task-queues/TRANSFER_MONEY_TASK_QUEUE).

## Workflow logic
The MoneyTransfer workflow in the appworkflow package orchestrates a money transfer process. The workflow involves the following steps:  
 * Create Transaction: 
   Initiates a transaction by calling the CreateTransaction activity.
 * Debit: 
   Debits the specified amount from the source account by calling the Debit activity.
 * Credit: 
  Credits the specified amount to the destination account by calling the Credit activity.
 * Compensation: 
   If any error occurs during the credit operation, the workflow compensates by rolling back the credit operation using the RollbackCredit activity.

## What Next?

You can run the Workflow code a few more times with `go run start/main.go` to understand how it interacts with the Worker and Temporal Server.

Please [read the tutorial](https://learn.temporal.io/getting_started/go/first_program_in_go/) for more details.
