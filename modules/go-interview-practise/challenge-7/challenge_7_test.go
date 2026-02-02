package challenge7

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestNewBankAccount(t *testing.T) {
	testCases := []struct {
		name           string
		id             string
		owner          string
		initialBalance float64
		minBalance     float64
		shouldError    bool
		errorType      string
	}{
		{
			name:           "Valid account creation",
			id:             "ACC001",
			owner:          "Alice",
			initialBalance: 1000.0,
			minBalance:     100.0,
			shouldError:    false,
		},
		{
			name:           "Empty ID",
			id:             "",
			owner:          "Alice",
			initialBalance: 1000.0,
			minBalance:     100.0,
			shouldError:    true,
			errorType:      "AccountError",
		},
		{
			name:           "Empty owner",
			id:             "ACC001",
			owner:          "",
			initialBalance: 1000.0,
			minBalance:     100.0,
			shouldError:    true,
			errorType:      "AccountError",
		},
		{
			name:           "Negative initial balance",
			id:             "ACC001",
			owner:          "Alice",
			initialBalance: -100.0,
			minBalance:     100.0,
			shouldError:    true,
			errorType:      "NegativeAmountError",
		},
		{
			name:           "Negative min balance",
			id:             "ACC001",
			owner:          "Alice",
			initialBalance: 1000.0,
			minBalance:     -100.0,
			shouldError:    true,
			errorType:      "NegativeAmountError",
		},
		{
			name:           "Initial balance less than min balance",
			id:             "ACC001",
			owner:          "Alice",
			initialBalance: 50.0,
			minBalance:     100.0,
			shouldError:    true,
			errorType:      "InsufficientFundsError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			account, err := NewBankAccount(tc.id, tc.owner, tc.initialBalance, tc.minBalance)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(fmt.Sprintf("%T", err), tc.errorType) {
					t.Errorf("Expected error of type %s but got %T", tc.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
					return
				}

				if account == nil {
					t.Errorf("Expected account to be created but it was nil")
					return
				}

				if account.ID != tc.id {
					t.Errorf("Expected ID %s but got %s", tc.id, account.ID)
				}

				if account.Owner != tc.owner {
					t.Errorf("Expected Owner %s but got %s", tc.owner, account.Owner)
				}

				if account.Balance != tc.initialBalance {
					t.Errorf("Expected Balance %.2f but got %.2f", tc.initialBalance, account.Balance)
				}

				if account.MinBalance != tc.minBalance {
					t.Errorf("Expected MinBalance %.2f but got %.2f", tc.minBalance, account.MinBalance)
				}
			}
		})
	}
}

func TestDeposit(t *testing.T) {
	testCases := []struct {
		name        string
		amount      float64
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid deposit",
			amount:      500.0,
			shouldError: false,
		},
		{
			name:        "Zero deposit",
			amount:      0.0,
			shouldError: false,
		},
		{
			name:        "Negative deposit",
			amount:      -100.0,
			shouldError: true,
			errorType:   "NegativeAmountError",
		},
		{
			name:        "Exceeds limit deposit",
			amount:      MaxTransactionAmount + 1.0,
			shouldError: true,
			errorType:   "ExceedsLimitError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			account, _ := NewBankAccount("TEST", "TestUser", 1000.0, 100.0)
			initialBalance := account.Balance

			err := account.Deposit(tc.amount)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(fmt.Sprintf("%T", err), tc.errorType) {
					t.Errorf("Expected error of type %s but got %T", tc.errorType, err)
				}

				// Balance should not change on error
				if account.Balance != initialBalance {
					t.Errorf("Expected balance to remain %.2f but got %.2f", initialBalance, account.Balance)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
					return
				}

				// Check if balance increased correctly
				expectedBalance := initialBalance + tc.amount
				if account.Balance != expectedBalance {
					t.Errorf("Expected balance %.2f but got %.2f", expectedBalance, account.Balance)
				}
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	testCases := []struct {
		name        string
		amount      float64
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid withdrawal",
			amount:      200.0,
			shouldError: false,
		},
		{
			name:        "Zero withdrawal",
			amount:      0.0,
			shouldError: false,
		},
		{
			name:        "Negative withdrawal",
			amount:      -100.0,
			shouldError: true,
			errorType:   "NegativeAmountError",
		},
		{
			name:        "Exceeds limit withdrawal",
			amount:      MaxTransactionAmount + 1.0,
			shouldError: true,
			errorType:   "ExceedsLimitError",
		},
		{
			name:        "Insufficient funds withdrawal",
			amount:      950.0, // Leaves 50, which is below min balance of 100
			shouldError: true,
			errorType:   "InsufficientFundsError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			account, _ := NewBankAccount("TEST", "TestUser", 1000.0, 100.0)
			initialBalance := account.Balance

			err := account.Withdraw(tc.amount)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(fmt.Sprintf("%T", err), tc.errorType) {
					t.Errorf("Expected error of type %s but got %T", tc.errorType, err)
				}

				// Balance should not change on error
				if account.Balance != initialBalance {
					t.Errorf("Expected balance to remain %.2f but got %.2f", initialBalance, account.Balance)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
					return
				}

				// Check if balance decreased correctly
				expectedBalance := initialBalance - tc.amount
				if account.Balance != expectedBalance {
					t.Errorf("Expected balance %.2f but got %.2f", expectedBalance, account.Balance)
				}
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	testCases := []struct {
		name        string
		amount      float64
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid transfer",
			amount:      300.0,
			shouldError: false,
		},
		{
			name:        "Zero transfer",
			amount:      0.0,
			shouldError: false,
		},
		{
			name:        "Negative transfer",
			amount:      -100.0,
			shouldError: true,
			errorType:   "NegativeAmountError",
		},
		{
			name:        "Exceeds limit transfer",
			amount:      MaxTransactionAmount + 1.0,
			shouldError: true,
			errorType:   "ExceedsLimitError",
		},
		{
			name:        "Insufficient funds transfer",
			amount:      950.0, // Leaves 50, which is below min balance of 100
			shouldError: true,
			errorType:   "InsufficientFundsError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			source, _ := NewBankAccount("SRC", "Source", 1000.0, 100.0)
			target, _ := NewBankAccount("TGT", "Target", 500.0, 50.0)

			sourceInitialBalance := source.Balance
			targetInitialBalance := target.Balance

			err := source.Transfer(tc.amount, target)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(fmt.Sprintf("%T", err), tc.errorType) {
					t.Errorf("Expected error of type %s but got %T", tc.errorType, err)
				}

				// Balances should not change on error
				if source.Balance != sourceInitialBalance {
					t.Errorf("Expected source balance to remain %.2f but got %.2f", sourceInitialBalance, source.Balance)
				}

				if target.Balance != targetInitialBalance {
					t.Errorf("Expected target balance to remain %.2f but got %.2f", targetInitialBalance, target.Balance)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
					return
				}

				// Check if balances changed correctly
				expectedSourceBalance := sourceInitialBalance - tc.amount
				expectedTargetBalance := targetInitialBalance + tc.amount

				if source.Balance != expectedSourceBalance {
					t.Errorf("Expected source balance %.2f but got %.2f", expectedSourceBalance, source.Balance)
				}

				if target.Balance != expectedTargetBalance {
					t.Errorf("Expected target balance %.2f but got %.2f", expectedTargetBalance, target.Balance)
				}
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	account, _ := NewBankAccount("CONC", "Concurrency Test", 1000.0, 100.0)

	const numOperations = 100
	var wg sync.WaitGroup
	wg.Add(numOperations * 2) // Add and withdraw operations

	// Perform concurrent deposits
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			_ = account.Deposit(10.0)
		}()
	}

	// Perform concurrent withdrawals
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			_ = account.Withdraw(5.0)
		}()
	}

	wg.Wait()

	// After all operations, the final balance should be:
	// Initial: 1000
	// 100 deposits of 10: +1000
	// 100 withdrawals of 5: -500
	// Expected: 1500
	expectedBalance := 1500.0

	if account.Balance != expectedBalance {
		t.Errorf("Expected balance after concurrent operations to be %.2f but got %.2f", expectedBalance, account.Balance)
	}
}
