package repositories

import (
	"fmt"
	"time"

	"dentvisor-backend/pkg/database"
	"github.com/google/uuid"
)

type Appointment struct {
	ID        uuid.UUID
	ClinicID  uuid.UUID
	PatientID uuid.UUID
	DoctorID  uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	Status    string
	Notes     string
}

type AppointmentRepository struct{}

func (r *AppointmentRepository) CreateAppointment(clinicID, patientID, doctorID uuid.UUID, startTime, endTime time.Time, status, notes string) error {
	// tsrange kullanımı: '[)' başlangıç dahil, bitiş hariç
	query := `
		INSERT INTO appointments (clinic_id, patient_id, doctor_id, time_range, status, notes) 
		VALUES (?, ?, ?, tsrange(?, ?, '[)'), ?, ?)
	`
	err := database.DB.Exec(query, clinicID, patientID, doctorID, startTime, endTime, status, notes).Error
	if err != nil {
		return fmt.Errorf("randevu oluşturulurken hata (çakışma olabilir): %w", err)
	}
	return nil
}
