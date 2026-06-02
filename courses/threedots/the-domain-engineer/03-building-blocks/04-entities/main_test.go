// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBankAccount_emptyID(t *testing.T) {
	a, err := NewBankAccount("", "Alice", nil)
	assert.Nil(t, a)
	assert.EqualError(t, err, "empty id")
}

func TestNewBankAccount_emptyOwner(t *testing.T) {
	a, err := NewBankAccount("A-1", "", nil)
	assert.Nil(t, a)
	assert.EqualError(t, err, "empty owner")
}

func TestNewBankAccount_zeroTransfer(t *testing.T) {
	valid, err := NewTransfer(100, time.Now())
	require.NoError(t, err)

	a, err := NewBankAccount("A-1", "Alice", []Transfer{valid, {}})
	assert.Nil(t, a)
	assert.EqualError(t, err, "1 transfer is zero")
}

func TestBankAccount_identity(t *testing.T) {
	a, err := NewBankAccount("A-1", "Alice", nil)
	require.NoError(t, err)

	b, err := NewBankAccount("A-2", "Alice", nil)
	require.NoError(t, err)

	assert.NotEqual(t, a.ID(), b.ID(), "two accounts with the same owner should still have distinct identities")
}

func TestBankAccount_Deposit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		a, err := NewBankAccount("A-1", "Alice", nil)
		require.NoError(t, err)

		assert.NoError(t, a.Deposit(100))
		assert.Equal(t, int64(100), a.Balance())
		assert.ElementsMatch(
			t,
			[]Transfer{
				MustNewTransfer(100, now),
			},
			a.Transfers(),
		)
	})
}

func TestBankAccount_Deposit_rejectsNonPositive(t *testing.T) {
	a, err := NewBankAccount("A-1", "Alice", nil)
	require.NoError(t, err)

	assert.EqualError(t, a.Deposit(0), "deposit amount must be positive")
	assert.EqualError(t, a.Deposit(-10), "deposit amount must be positive")
	assert.Empty(t, a.Transfers())
}

func TestBankAccount_Withdraw(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		a, err := NewBankAccount("A-1", "Alice", nil)
		require.NoError(t, err)
		require.NoError(t, a.Deposit(100))

		assert.NoError(t, a.Withdraw(40))
		assert.Equal(t, int64(60), a.Balance())
		assert.ElementsMatch(
			t,
			[]Transfer{
				MustNewTransfer(100, now),
				MustNewTransfer(-40, now),
			},
			a.Transfers(),
		)
	})
}

func TestBankAccount_Withdraw_rejectsNonPositive(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		a, err := NewBankAccount("A-1", "Alice", nil)
		require.NoError(t, err)
		require.NoError(t, a.Deposit(100))

		assert.EqualError(t, a.Withdraw(0), "withdraw amount must be positive")
		assert.EqualError(t, a.Withdraw(-10), "withdraw amount must be positive")
		assert.ElementsMatch(
			t,
			[]Transfer{
				MustNewTransfer(100, now),
			},
			a.Transfers(),
		)
	})
}

func TestBankAccount_Withdraw_preventsOverdraft(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		a, err := NewBankAccount("A-1", "Alice", nil)
		require.NoError(t, err)
		require.NoError(t, a.Deposit(50))

		assert.EqualError(t, a.Withdraw(100), "insufficient funds")
		assert.Equal(t, int64(50), a.Balance(), "balance should not change after a failed withdraw")
		assert.ElementsMatch(
			t,
			[]Transfer{
				MustNewTransfer(50, now),
			},
			a.Transfers(),
		)
	})
}

func TestBankAccount_Transfers_returnsCopy(t *testing.T) {
	a, err := NewBankAccount("A-1", "Alice", nil)
	require.NoError(t, err)
	require.NoError(t, a.Deposit(100))

	transfers := a.Transfers()
	transfers[0] = Transfer{}

	assert.False(t, a.Transfers()[0].IsZero(), "Transfers() should return a copy, mutating it should not affect the account")
}
