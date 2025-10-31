package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func newAppointmentServiceWithMock(t *testing.T) (*AppointmentService, sqlmock.Sqlmock) {
	t.Helper()

	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() {
		_ = bunDB.Close()
	})

	return NewAppointmentService(bunDB), mock
}

func TestListAppointments_FilterByClientID(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("1", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{ClientID: "client123"}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "client123", result[0].ClientID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByProviderID(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("2", "client999", "provider456", "West", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{ProviderID: "provider456"}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "provider456", result[0].ProviderID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByBranch(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("3", "client111", "provider222", "East", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{Branch: "East"}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "East", result[0].Branch)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByDate(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	targetDate := time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("4", "client555", "provider666", "North", targetDate, targetDate.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{Date: "2025-10-20"}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_InvalidDateFormat(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"})

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{Date: "invalid-date"}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_MultipleFilters(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("5", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
	}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_NoFilters(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("6", "client1", "provider1", "Main", now, now.Add(time.Hour), "scheduled", nil).
		AddRow("7", "client2", "provider2", "East", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_QueryError(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT count").
		WillReturnError(errors.New("database connection lost"))

	filter := models.AppointmentFilter{}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "count appointments")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_ScanError(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("1", "client123", "provider456", "Main", "invalid", time.Now(), "scheduled", nil)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, total, err := svc.ListAppointments(ctx, filter)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "query appointments")
	}
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_EmptyResult(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"})

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_WithNotes(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	notes := "Important appointment"
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("8", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", notes)

	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, total, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	assert.NotNil(t, result[0].Notes)
	assert.Equal(t, notes, *result[0].Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_Success(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	notes := "Test notes"
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "client456", "provider789", "Main", now, now.Add(time.Hour), "scheduled", notes)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "appt123", result.ID)
	assert.Equal(t, "client456", result.ClientID)
	assert.Equal(t, "provider789", result.ProviderID)
	assert.Equal(t, "Main", result.Branch)
	assert.Equal(t, "scheduled", result.Status)
	assert.NotNil(t, result.Notes)
	assert.Equal(t, notes, *result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_NotFound(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	result, err := svc.GetAppointmentByID(ctx, "nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_DatabaseError(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").
		WillReturnError(errors.New("database failure"))

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query appointment by id")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_WithoutNotes(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "client456", "provider789", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_SuccessWithNotes(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	notes := "Important appointment"
	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      &notes,
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	// Provider availability check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := svc.CreateAppointment(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "client123", result.ClientID)
	assert.Equal(t, "provider456", result.ProviderID)
	assert.Equal(t, "Main", result.Branch)
	assert.Equal(t, "scheduled", result.Status)
	require.NotNil(t, result.Notes)
	assert.Equal(t, notes, *result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_SuccessWithoutNotes(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	// Provider availability check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := svc.CreateAppointment(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Nil(t, result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_Conflict(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	// Provider availability check - conflict
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectRollback()

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAppointmentConflict)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_ClientConflict(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check - conflict
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectRollback()

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrClientAlreadyBooked)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_BeginTransactionFailure(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin().WillReturnError(errors.New("transaction start failed"))

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_InsertFailure(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	// Provider availability check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectExec("INSERT").WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insert appointment")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_CommitFailure(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	// Client existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Provider existence check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Client already booked check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	// Provider availability check
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_StatusChange(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	cancelled := models.StatusCancelled
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := models.UpdateAppointmentRequest{ClientID: "client123", Status: &cancelled}
	result, err := svc.UpdateAppointment(ctx, "appt123", req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "appt123", result.ID)
	assert.Equal(t, models.StatusCancelled, result.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_NotFound(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	req := models.UpdateAppointmentRequest{ClientID: "client123"}
	_, err := svc.UpdateAppointment(ctx, "missing", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAppointmentNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_Forbidden(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "other-client", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectRollback()

	req := models.UpdateAppointmentRequest{ClientID: "client123"}
	_, err := svc.UpdateAppointment(ctx, "appt123", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAppointmentForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_InvalidStatus(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	invalidStatus := "invalid"
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectRollback()

	req := models.UpdateAppointmentRequest{ClientID: "client123", Status: &invalidStatus}
	_, err := svc.UpdateAppointment(ctx, "appt123", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStatus)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_UpdateFailure(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	now := time.Now()
	notes := "Updated notes"
	rows := sqlmock.NewRows([]string{"id", "clientid", "providerid", "branch", "starttime", "endtime", "status", "notes"}).
		AddRow("appt123", "client123", "provider456", "Main", now, now.Add(time.Hour), "scheduled", nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectExec("UPDATE").WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	req := models.UpdateAppointmentRequest{ClientID: "client123", Notes: &notes}
	_, err := svc.UpdateAppointment(ctx, "appt123", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update appointment")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAppointment_SelectFailure(t *testing.T) {
	svc, mock := newAppointmentServiceWithMock(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("select failed"))
	mock.ExpectRollback()

	req := models.UpdateAppointmentRequest{ClientID: "client123"}
	_, err := svc.UpdateAppointment(ctx, "appt123", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update appointment")
	assert.NoError(t, mock.ExpectationsWereMet())
}
