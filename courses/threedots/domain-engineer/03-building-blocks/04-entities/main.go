package main

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// BankAccount is an entity: it has an identity (ID), and its state (balance)
// only changes through methods that enforce the invariants.
//
// Turn this into a proper entity:
//  1. Reject invalid inputs in NewBankAccount (empty id, empty owner, zero transfer).
//  2. Make Deposit reject non-positive amounts.
//  3. Make Withdraw reject non-positive amounts and prevent overdraft.
//
// Encapsulation (unexported fields) is already in place. The tests in
// main_test.go describe the target behavior.
type BankAccount struct {
	id        string
	owner     string
	balance   int64
	transfers []Transfer
}

func NewBankAccount(id, owner string, transfers []Transfer) (*BankAccount, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if owner == "" {
		return nil, errors.New("empty owner")
	}
	for i, t := range transfers {
		if t.IsZero() {
			return nil, fmt.Errorf("%d transfer is zero", i)
		}
	}

	return &BankAccount{
		id:        id,
		owner:     owner,
		transfers: transfers,
	}, nil
}

func (a *BankAccount) ID() string {
	return a.id
}

func (a *BankAccount) Owner() string {
	return a.owner
}

func (a *BankAccount) Balance() int64 {
	return a.balance
}

func (a *BankAccount) Deposit(amount int64) error {
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	a.balance += amount

	a.transfers = append(a.transfers, Transfer{
		when:   time.Now(),
		amount: amount,
	})

	return nil
}

func (a *BankAccount) Withdraw(amount int64) error {
	if amount <= 0 {
		return errors.New("withdraw amount must be positive")
	}
	if amount > a.balance {
		return errors.New("insufficient funds")
	}

	a.balance -= amount

	a.transfers = append(a.transfers, Transfer{
		when:   time.Now(),
		amount: -amount,
	})

	return nil
}

func (a *BankAccount) Transfers() []Transfer {
	// Returning a copy, so such modification is not possible:
	//  transfers := a.Transfers()
	//	transfers[0] = Transfer{}
	return slices.Clone(a.transfers)
}

type Transfer struct {
	when   time.Time
	amount int64
}

func NewTransfer(amount int64, when time.Time) (Transfer, error) {
	if when.IsZero() {
		return Transfer{}, errors.New("when is zero")
	}

	return Transfer{amount: amount, when: when}, nil
}

func MustNewTransfer(amount int64, when time.Time) Transfer {
	t, err := NewTransfer(amount, when)
	if err != nil {
		panic(err)
	}
	return t
}

func (t Transfer) When() time.Time {
	return t.when
}

func (t Transfer) Amount() int64 {
	return t.amount
}

func (t Transfer) IsZero() bool {
	return t == Transfer{}
}

func main() {
	acc, _ := NewBankAccount("A-1", "Alice", nil)
	_ = acc.Deposit(100)
	_ = acc.Withdraw(30)
	fmt.Printf("%s balance: %d\n", acc.Owner(), acc.Balance())
}
