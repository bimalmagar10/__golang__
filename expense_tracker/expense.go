package expense

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Expense struct {
	ID int   `json:"id"`
	Amount float64 `json:"amount"`
	Description string `json:"description"`
	Date time.Time `json:"date"`
}

type Expenses struct {
	Expenses []Expense `json:"expenses"`
	NextID int `json:"next_id"`
}

type ExpenseStore struct {
	filePath string
	expenses Expenses
}

func NewExpenseStore(filepath string) (*ExpenseStore,error) {
	store := &ExpenseStore{
		filePath: filepath,
		expenses: Expenses{
			Expenses: []Expense{},
		},
	}

	//Check if the filepath exists in the current directory
	if _,err := os.Stat(filepath); err == nil {
		data,err := os.ReadFile(filepath)
		if err != nil {
			return nil,fmt.Errorf("failed to read the file:%v",err)
		}

		if len(data) > 0 {
			if err:= json.Unmarshal(data,&store.expenses); err != nil {
				return nil,fmt.Errorf("failed to parse expense data: %v",err)
			}
		}

	}

	return store,nil
}

func (store *ExpenseStore) Save() error {
	data,err := json.MarshalIndent(store.expenses,""," ")

	if err != nil {
		return fmt.Errorf("Failed to convert to json: %v",err)
	}

	if err := os.WriteFile(store.filePath,data,0644); err != nil {
		return fmt.Errorf("Failes to write expenses to JSON file %v",err)
	}

	return nil
}


func (store *ExpenseStore) AddExpense(description string, amount float64) (int,error) {
	if amount <=0 {
		return 0,errors.New("Amount must be positive \n")
	}

	if description == "" {
		return 0, errors.New("Description cannot be empty!")
	}
	expense :=  Expense{}

	if len(store.expenses.Expenses) == 0 {
		expense = Expense {
			ID: 1,
			Description: description,
			Amount: amount,
			Date: time.Now(),
		}
		store.expenses.NextID = 1
	} else {
		expense = Expense {
			ID: store.expenses.NextID,
			Description: description,
			Amount: amount,
			Date: time.Now(),
		}
	}
	
	store.expenses.Expenses = append(store.expenses.Expenses,expense)
	store.expenses.NextID++


	if err:= store.Save(); err != nil {
		return 0, err
	}

	return expense.ID,nil
	
}

func (store *ExpenseStore) UpdateExpense(id int, description string,amount float64) error {
	found := false
	
	for idx,expense := range store.expenses.Expenses {
		if expense.ID == id {
			if amount <= 0 {
				return fmt.Errorf("Amount must be positive and non zero")
			}

			if description != "" {
				store.expenses.Expenses[idx].Description = description
			}

			store.expenses.Expenses[idx].Amount = amount

			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Expense with ID (%v) not found \n",id)
	}

	return store.Save()
}

func (store *ExpenseStore) DeleteExpense(id int) error{
	found := false
	
	for idx,expense:= range store.expenses.Expenses {
		if expense.ID == id {
			store.expenses.Expenses = append(store.expenses.Expenses[:idx],store.expenses.Expenses[idx+1:]...)
			found =true
			break
		}
	}

	if !found {
		return fmt.Errorf("Error: Expense with id %v not found",id)
	}

	return store.Save()
}

func (store *ExpenseStore) GetExpenses() []Expense {
	return store.expenses.Expenses
}

func (store *ExpenseStore) GetMonthlyExpenses(month int) (float64,error) {
	if month < 1 || month > 12 {
		return 0, errors.New("Month must be between 1 and 12 \n")
	}

	currentYear := time.Now().Year()

	total := 0.0

	for _,expense := range store.expenses.Expenses {
		if expense.Date.Month() == time.Month(month) && currentYear == expense.Date.Year() {
			total += expense.Amount
		}
	}

	return total,nil
}

func (store *ExpenseStore) GetTotalExpenses() (float64,error) {
	if len(store.expenses.Expenses) == 0 {
		return 0,errors.New("No expenses recorded \n")
	}

	total := 0.0

	for _,expense:= range store.expenses.Expenses {
		total += expense.Amount
	}

	return total,nil
}