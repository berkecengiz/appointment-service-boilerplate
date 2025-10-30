package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func newProviderServiceWithMock(t *testing.T) (*ProviderService, sqlmock.Sqlmock) {
	t.Helper()

	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() {
		_ = bunDB.Close()
	})

	return NewProviderService(bunDB), mock
}

func TestCreateProvider_Success(t *testing.T) {
	svc, mock := newProviderServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectExec("INSERT").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.CreateProvider(ctx, models.CreateProviderRequest{
		Name:    "Dr. Smith",
		Email:   "smith@example.com",
		Phone:   "+1-555-0202",
		Address: "456 Wellness Ave",
	})

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Dr. Smith", result.Name)
	_, parseErr := uuid.Parse(result.ID)
	assert.NoError(t, parseErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProvider_UniqueEmail(t *testing.T) {
	svc, mock := newProviderServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectExec("INSERT").
		WillReturnError(errors.New("duplicate key value violates unique constraint \"providers_email_key\""))

	_, err := svc.CreateProvider(ctx, models.CreateProviderRequest{
		Name:    "Dr. Smith",
		Email:   "smith@example.com",
		Phone:   "+1-555-0202",
		Address: "456 Wellness Ave",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderEmailExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProviders_OrderByName(t *testing.T) {
	svc, mock := newProviderServiceWithMock(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "address"}).
		AddRow("1", "Alpha", "alpha@example.com", "111", "A St").
		AddRow("2", "Beta", "beta@example.com", "222", "B St")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.ListProviders(ctx)

	assert.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Alpha", result[0].Name)
	assert.Equal(t, "Beta", result[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProviderByID_NotFound(t *testing.T) {
	svc, mock := newProviderServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	result, err := svc.GetProviderByID(ctx, "missing")

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProviderByID_Success(t *testing.T) {
	svc, mock := newProviderServiceWithMock(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "address"}).
		AddRow("1", "Alpha", "alpha@example.com", "111", "A St")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.GetProviderByID(ctx, "1")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Alpha", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
