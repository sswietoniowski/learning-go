// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/adapters/db"
	"eats/backend/billing/adapters/db/dbtests"
	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/testutils"
)

func TestCreateDocument_ConcurrentDocumentNumbers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := testutils.NewDB(t)

	repo := db.NewPostgresRepository(database)

	seriesStr := common.NewUUIDv7().String()
	series, err := domain.NewDocumentSeries(seriesStr)
	require.NoError(t, err)

	q := dbtests.New(database)
	err = q.SaveDocumentSeries(ctx, seriesStr)
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	concurrentDocuments := 100

	for i := 0; i < concurrentDocuments; i++ {
		wg.Go(func() {
			_, err := repo.CreateDocument(
				ctx,
				series,
				func(docNumber domain.DocumentNumber) (db.DocumentRecord, error) {
					return db.DocumentRecord{
						UUID: domain.DocumentUUID{UUID: common.NewUUIDv7()},
					}, nil
				},
			)
			assert.NoError(t, err)
		})
	}

	wg.Wait()

	docs, err := q.GetDocumentsBySeriesPrefix(ctx, series.String())
	require.NoError(t, err)

	require.Len(t, docs, concurrentDocuments)

	// Assert no sequence gaps and no duplicates
	for i := 0; i < concurrentDocuments; i++ {
		doc := docs[i]
		docNumber := strings.ReplaceAll(doc, seriesStr+"-", "")
		number, err := strconv.ParseInt(docNumber, 10, 64)
		assert.NoError(t, err)

		assert.Equal(t, i+1, int(number))
	}
}
