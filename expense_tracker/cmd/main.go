package main

import (
	expense "expense_tracker"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
) // We call this factored import in Golang

func main() {
	fmt.Println("Welcome to the expense Bimal's tracker application!\n")
	currentDir , error := os.Getwd()

	if error != nil {
		fmt.Fprint(os.Stderr,"Errir getting current directory: %v\n",error)
		os.Exit(1)
	}

	outDir := filepath.Join(currentDir,"data")
	if error := os.MkdirAll(outDir,0755); error != nil {
		fmt.Fprint(os.Stderr,"Error creating data directory: %v \n",error)
	}

	out:= filepath.Join(outDir,"expenses.json")
	
	store,err := expense.NewExpenseStore(out)
	
	if err != nil {
		fmt.Fprint(os.Stderr,"Error initializing expense store: %v \n",err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add":
		handleAdd(store)
	case "update":
		handleUpdate(store)
	case "delete":
		handleDelete(store)
	case "list":
		handleList(store)
	case "summary":
		handleSummary(store)
	default:
		fmt.Fprintf(os.Stderr,"Unknown command %s \n",command)
		os.Exit(1)
	}

}


func handleAdd(store *expense.ExpenseStore) {

	// Setup the command ling flags
	addCmd := flag.NewFlagSet("add",flag.ExitOnError)
	description := addCmd.String("description","","Description of expense")
	amount := addCmd.Float64("amount",0.0,"Amount of the expense")

	addCmd.Parse(os.Args[2:])

	if *description == "" || *amount <=0 {
		fmt.Fprintf(os.Stderr,"Error: Please enter both description and amount greater than 0!\n")
		os.Exit(1)
	}

	// Add the expenses
	expenseID, err := store.AddExpense(*description,*amount)
	if err != nil {
		fmt.Fprint(os.Stderr,"Error adding expense: %v \n",err)
		os.Exit(1)
	}

	fmt.Printf("# Expense added successfully (ID: %d) \n",expenseID)
}

func handleUpdate(store *expense.ExpenseStore) {

	updateCmd := flag.NewFlagSet("update",flag.ExitOnError)
	id := updateCmd.Int("id",0,"ID of the expense to update")
	description := updateCmd.String("description","","New description to update")
	amount := updateCmd.Float64("amount",0.0,"New Amount")

	updateCmd.Parse(os.Args[2:])

	if *id <=0 || *amount <=0 {
		fmt.Fprint(os.Stderr,"Error: amount and id should not be 0 \n")
		os.Exit(1)
	}


	if err := store.UpdateExpense(*id,*description,*amount); err != nil {
		fmt.Fprint(os.Stderr,"Error updating expense: %v \n",err)
		os.Exit(1)
	}

	fmt.Println("# Expense updated sucessfully! \n")
}

func handleDelete(store *expense.ExpenseStore) {

	deleteCmd := flag.NewFlagSet("delete",flag.ExitOnError)
	id := deleteCmd.Int("id",0,"ID of the expense to delete")

	deleteCmd.Parse(os.Args[2:])

	if *id <= 0 {
		fmt.Fprint(os.Stderr,"Error: id must not be negative or 0 \n")
		os.Exit(1)
	}

	if err:= store.DeleteExpense(*id);err != nil {
		fmt.Fprint(os.Stderr,"Error when deleting expense: %v \n",err)
		os.Exit(1)
	}

	fmt.Println("# Expense deleted successfully! \n")
}

func handleList(store *expense.ExpenseStore) {
	expenses := store.GetExpenses()

	if len(expenses) == 0 {
		fmt.Println("# You do not have any expenses.Please add to list all your expenses!")
		return
	}
	//Print headers
	fmt.Printf("# %-2s| %-10s| %-30s| %-10s \n","ID","Date","Description","Amount")
	fmt.Printf("===============================================================\n")

	for _,expense:= range expenses {
		dateStr := expense.Date.Format("2006-01-02")
		fmt.Printf("# %-2d| %-10s| %-30s| $%-10.2f \n",expense.ID,dateStr,expense.Description,expense.Amount)
	}
}

func handleSummary(store *expense.ExpenseStore) {
	if len(os.Args) > 2 && os.Args[2] == "--month" {
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr,"Error: Month value is required \n")
			os.Exit(1)
		}

		month,err := strconv.Atoi(os.Args[3])
		if err != nil || month < 1 || month > 12 {
			fmt.Fprintf(os.Stderr,"Error: month must be between 1 and 12 \n")
			os.Exit(1)
		}

		total,err := store.GetMonthlyExpenses(month)

		if err != nil {
			fmt.Fprintf(os.Stderr,"Error getting monthly expenses:%v \n",err)
		}

		monthName := time.Month(month).String()

		fmt.Printf("# Total expenses for %s:$ %.2f \n",monthName,total)
	} else {
		total,err := store.GetTotalExpenses()
		if err != nil {
			fmt.Fprintf(os.Stderr,"Error getting total expenses: %v \n",err)
			os.Exit(1)
		}
		fmt.Printf("# Total expenses: $%.2f \n",total)
	}
}