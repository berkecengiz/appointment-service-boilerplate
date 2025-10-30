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

func newClientServiceWithMock(t *testing.T) (*ClientService, sqlmock.Sqlmock) {
	t.Helper()

	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() {
		_ = bunDB.Close()
	})

	return NewClientService(bunDB), mock
}

func TestCreateClient_Success(t *testing.T) {
	svc, mock := newClientServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectExec("INSERT").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.CreateClient(ctx, models.CreateClientRequest{
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		Phone:   "+1-555-0101",
		Address: "123 Market St",
	})

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Jane Doe", result.Name)
	assert.Equal(t, "jane@example.com", result.Email)
	_, parseErr := uuid.Parse(result.ID)
	assert.NoError(t, parseErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateClient_UniqueEmail(t *testing.T) {
	svc, mock := newClientServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectExec("INSERT").
		WillReturnError(errors.New("duplicate key value violates unique constraint \"clients_email_key\""))

	_, err := svc.CreateClient(ctx, models.CreateClientRequest{
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		Phone:   "+1-555-0101",
		Address: "123 Market St",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrClientEmailExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListClients_OrderByName(t *testing.T) {
	svc, mock := newClientServiceWithMock(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "address"}).
		AddRow("1", "Alice", "alice@example.com", "111", "A St").
		AddRow("2", "Bob", "bob@example.com", "222", "B St")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.ListClients(ctx)

	assert.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Alice", result[0].Name)
	assert.Equal(t, "Bob", result[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetClientByID_NotFound(t *testing.T) {
	svc, mock := newClientServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	result, err := svc.GetClientByID(ctx, "missing")

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetClientByID_Success(t *testing.T) {
	svc, mock := newClientServiceWithMock(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "address"}).
		AddRow("1", "Alice", "alice@example.com", "111", "A St")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.GetClientByID(ctx, "1")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Alice", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
