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
)

func TestListAppointments_FilterByCustomerID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("1", "customer123", "provider456", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE CustomerId = \$1 ORDER BY StartTime ASC`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := models.AppointmentFilter{CustomerID: "customer123"}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "customer123", result[0].CustomerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByProviderID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("2", "customer999", "provider456", "West", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE ProviderId = \$1 ORDER BY StartTime ASC`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := models.AppointmentFilter{ProviderID: "provider456"}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "provider456", result[0].ProviderID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByBranch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("3", "customer111", "provider222", "East", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Branch = \$1 ORDER BY StartTime ASC`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := models.AppointmentFilter{Branch: "East"}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "East", result[0].Branch)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_FilterByDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	targetDate := time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("4", "customer555", "provider666", "North", targetDate, targetDate.Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE StartTime >= \$1 AND StartTime < \$2 ORDER BY StartTime ASC`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := models.AppointmentFilter{Date: "2025-10-20"}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_InvalidDateFormat(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	// Invalid date format should be gracefully ignored
	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"})

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnRows(rows)

	filter := models.AppointmentFilter{Date: "invalid-date"}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_MultipleFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("5", "customer123", "provider456", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE CustomerId = \$1 AND ProviderId = \$2 AND Branch = \$3 ORDER BY StartTime ASC`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := models.AppointmentFilter{
		CustomerID: "customer123",
		ProviderID: "provider456",
		Branch:     "Main",
	}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_NoFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("6", "customer1", "provider1", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", nil).
		AddRow("7", "customer2", "provider2", "East", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnError(errors.New("database connection lost"))

	filter := models.AppointmentFilter{}
	result, err := svc.ListAppointments(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query appointments")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	// Wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"Id", "CustomerId"}).
		AddRow("1", "customer123")

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, err := svc.ListAppointments(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "scan appointment")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"})

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAppointments_WithNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	notes := "Important appointment"
	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("8", "customer123", "provider456", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", notes)

	mock.ExpectQuery("SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments ORDER BY StartTime ASC").
		WillReturnRows(rows)

	filter := models.AppointmentFilter{}
	result, err := svc.ListAppointments(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NotNil(t, result[0].Notes)
	assert.Equal(t, notes, *result[0].Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	notes := "Test notes"
	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("appt123", "customer456", "provider789", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", notes)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "appt123", result.ID)
	assert.Equal(t, "customer456", result.CustomerID)
	assert.Equal(t, "provider789", result.ProviderID)
	assert.Equal(t, "Main", result.Branch)
	assert.Equal(t, "scheduled", result.Status)
	assert.NotNil(t, result.Notes)
	assert.Equal(t, notes, *result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	result, err := svc.GetAppointmentByID(ctx, "nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("database failure"))

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query appointment by id")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppointmentByID_WithoutNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "CustomerId", "ProviderId", "Branch", "StartTime", "EndTime", "Status", "Notes"}).
		AddRow("appt123", "customer456", "provider789", "Main", time.Now(), time.Now().Add(time.Hour), "scheduled", nil)

	mock.ExpectQuery(`SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	result, err := svc.GetAppointmentByID(ctx, "appt123")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_SuccessWithNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	notes := "Important appointment"
	req := models.CreateAppointmentRequest{
		CustomerID: "customer123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      &notes,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO Appointments").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow("new-appt-id"))
	mock.ExpectCommit()

	result, err := svc.CreateAppointment(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "new-appt-id", result.ID)
	assert.Equal(t, "customer123", result.CustomerID)
	assert.Equal(t, "provider456", result.ProviderID)
	assert.Equal(t, "Main", result.Branch)
	assert.Equal(t, "scheduled", result.Status)
	assert.NotNil(t, result.Notes)
	assert.Equal(t, notes, *result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_SuccessWithoutNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		CustomerID: "customer123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      nil,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO Appointments").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow("new-appt-id"))
	mock.ExpectCommit()

	result, err := svc.CreateAppointment(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "new-appt-id", result.ID)
	assert.Nil(t, result.Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_BeginTransactionFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		CustomerID: "customer123",
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
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		CustomerID: "customer123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO Appointments").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insert appointment")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAppointment_CommitFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAppointmentService(db)
	ctx := context.Background()

	req := models.CreateAppointmentRequest{
		CustomerID: "customer123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO Appointments").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow("new-appt-id"))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	result, err := svc.CreateAppointment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
