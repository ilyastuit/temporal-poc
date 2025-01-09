package app

const MoneyTransferTaskQueueName = "TRANSFER_MONEY_TASK_QUEUE"

type MoneyTransferWorkflowDetails struct {
	SourceAccount   string
	TargetAccount   string
	Amount          int
	TransactionType string
}
