package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/services"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CreateBookingParams represents the HTTP request body for creating a booking
type CreateBookingParams struct {
	ChildID        int64                         `json:"child_id" validate:"required,gt=0"`
	ServiceID      int64                         `json:"service_id" validate:"required,gt=0"`
	BookingType    string                        `json:"booking_type" validate:"required,oneof=trial single package_4 package_8 package_12"`
	SessionDates   []CreateBookingSessionParams  `json:"session_dates" validate:"required,min=1"`
	PreferredCoach *int64                        `json:"preferred_coach,omitempty"`
	ParentNotes    *string                       `json:"parent_notes,omitempty" validate:"omitempty,max=500"`
}

// CreateBookingSessionParams represents a session in the booking request
type CreateBookingSessionParams struct {
	ScheduleID  int64  `json:"schedule_id" validate:"required,gt=0"`
	SessionDate string `json:"session_date" validate:"required"` // YYYY-MM-DD format
}

// NewCreateBookingParams parses and validates booking creation request
func NewCreateBookingParams(r *http.Request) (*CreateBookingParams, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, internal.NewValidationError("Failed to read request body")
	}
	defer r.Body.Close()

	var params CreateBookingParams
	if err := json.Unmarshal(body, &params); err != nil {
		return nil, internal.NewValidationError("Invalid JSON format")
	}

	return &params, nil
}

// Validate validates the booking parameters
func (p *CreateBookingParams) Validate(ctx context.Context) error {
	if err := common.ValidateStruct(p); err != nil {
		return err
	}

	// Validate booking type matches session count
	bookingType := services.BookingType(p.BookingType)
	expectedSessions := bookingType.GetSessionCount()
	if len(p.SessionDates) != expectedSessions {
		return internal.NewValidationError(
			formatString("booking_type %s requires %d sessions, got %d", p.BookingType, expectedSessions, len(p.SessionDates)),
		)
	}

	// Validate each session date format
	for i, session := range p.SessionDates {
		if _, err := time.Parse("2006-01-02", session.SessionDate); err != nil {
			return internal.NewValidationError(
				formatString("session %d: invalid date format, expected YYYY-MM-DD", i+1),
			)
		}
	}

	return nil
}

// ToCreateBookingRequest converts DTO to domain request
func (p *CreateBookingParams) ToCreateBookingRequest(parentID int64) (*services.CreateBookingRequest, error) {
	sessionDates := make([]services.BookingSessionRequest, 0, len(p.SessionDates))

	for _, session := range p.SessionDates {
		sessionDate, err := time.Parse("2006-01-02", session.SessionDate)
		if err != nil {
			return nil, err
		}

		sessionDates = append(sessionDates, services.BookingSessionRequest{
			ScheduleID:  session.ScheduleID,
			SessionDate: sessionDate,
		})
	}

	return &services.CreateBookingRequest{
		ParentID:       parentID,
		ChildID:        p.ChildID,
		ServiceID:      p.ServiceID,
		BookingType:    services.BookingType(p.BookingType),
		SessionDates:   sessionDates,
		PreferredCoach: p.PreferredCoach,
		ParentNotes:    p.ParentNotes,
	}, nil
}

// BookingConfirmationResponse is a temporary response structure
type BookingConfirmationResponse struct {
	Data BookingConfirmationData `json:"data"`
}

type BookingConfirmationData struct {
	BookingID     int64                  `json:"booking_id"`
	BookingNumber string                 `json:"booking_number"`
	ServiceName   string                 `json:"service_name,omitempty"`
	ChildName     string                 `json:"child_name,omitempty"`
	BookingType   string                 `json:"booking_type"`
	TotalSessions int                    `json:"total_sessions"`
	TotalAmount   float64                `json:"total_amount"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	Sessions      []BookingSessionInfoDTO `json:"sessions,omitempty"`
}

type BookingSessionInfoDTO struct {
	SessionDate time.Time `json:"session_date"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	CoachName   *string   `json:"coach_name,omitempty"`
}

// ToV1BookingConfirmation converts domain confirmation to response
func ToV1BookingConfirmation(confirmation *services.BookingConfirmation) *BookingConfirmationResponse {
	sessions := make([]BookingSessionInfoDTO, 0, len(confirmation.Sessions))
	for _, s := range confirmation.Sessions {
		sessions = append(sessions, BookingSessionInfoDTO{
			SessionDate: s.SessionDate,
			StartTime:   s.StartTime,
			EndTime:     s.EndTime,
			CoachName:   s.CoachName,
		})
	}

	return &BookingConfirmationResponse{
		Data: BookingConfirmationData{
			BookingID:     confirmation.BookingID,
			BookingNumber: confirmation.BookingNumber,
			ServiceName:   confirmation.ServiceName,
			ChildName:     confirmation.ChildName,
			BookingType:   string(confirmation.BookingType),
			TotalSessions: confirmation.TotalSessions,
			TotalAmount:   confirmation.TotalAmount,
			Status:        string(confirmation.Status),
			CreatedAt:     confirmation.CreatedAt,
			Sessions:      sessions,
		},
	}
}

// formatString is a helper to format strings
func formatString(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// ToV1BookingDetail converts domain booking to OpenAPI v1 response
func ToV1BookingDetail(booking *services.Booking, enrichment *BookingEnrichment) *v1.BookingDetailResponse {
	// Convert booking type
	bookingTypeStr := string(booking.BookingType)
	bookingType := v1.BookingType(bookingTypeStr)

	// Convert status
	statusStr := string(booking.Status)
	status := v1.BookingStatus(statusStr)

	// Build sessions
	var sessions []v1.BookingSession
	completedCount := 0
	for i, session := range booking.Sessions {
		sessionStatusStr := string(session.Status)
		sessionStatus := v1.SessionStatus(sessionStatusStr)
		sessionNumber := i + 1

		// Convert time.Time to openapi_types.Date
		sessionDate := openapi_types.Date{Time: session.SessionDate}

		bookingSession := v1.BookingSession{
			Id:            &session.ID,
			SessionNumber: &sessionNumber,
			SessionDate:   &sessionDate,
			StartTime:     &session.StartTime,
			EndTime:       &session.EndTime,
			Status:        &sessionStatus,
		}

		if session.Status == services.SessionStatusCompleted {
			completedCount++
		}

		sessions = append(sessions, bookingSession)
	}

	detail := v1.BookingDetail{
		Id:                &booking.ID,
		BookingNumber:     &booking.BookingNumber,
		BookingType:       &bookingType,
		TotalSessions:     &booking.TotalSessions,
		TotalAmount:       &booking.TotalAmount,
		CompletedSessions: &completedCount,
		Status:            &status,
		ParentNotes:       booking.ParentNotes,
		CreatedAt:         &booking.CreatedAt,
		Sessions:          &sessions,
	}

	// Add enrichment data if available
	if enrichment != nil {
		if enrichment.ServiceName != "" {
			detail.Service = &struct {
				Category *string `json:"category,omitempty"`
				Id       *int64  `json:"id,omitempty"`
				Name     *string `json:"name,omitempty"`
			}{
				Id:   &booking.ServiceID,
				Name: &enrichment.ServiceName,
			}
			if enrichment.CategoryName != "" {
				detail.Service.Category = &enrichment.CategoryName
			}
		}

		if enrichment.ChildName != "" {
			detail.Child = &v1.Child{
				Id:   &booking.ChildID,
				Name: &enrichment.ChildName,
			}
		}

		if enrichment.VendorName != "" {
			detail.Vendor = &struct {
				BusinessName *string `json:"business_name,omitempty"`
				Id           *int64  `json:"id,omitempty"`
				Logo         *string `json:"logo,omitempty"`
			}{
				Id:           &booking.VendorID,
				BusinessName: &enrichment.VendorName,
			}
		}

		// Find next upcoming session
		now := time.Now()
		for _, session := range booking.Sessions {
			if session.SessionDate.After(now) && session.Status == services.SessionStatusScheduled {
				sessionDate := openapi_types.Date{Time: session.SessionDate}
				detail.NextSession = &struct {
					EndTime     *string             `json:"end_time,omitempty"`
					SessionDate *openapi_types.Date `json:"session_date,omitempty"`
					StartTime   *string             `json:"start_time,omitempty"`
				}{
					SessionDate: &sessionDate,
					StartTime:   &session.StartTime,
					EndTime:     &session.EndTime,
				}
				break
			}
		}
	}

	return &v1.BookingDetailResponse{
		Data: detail,
	}
}

// BookingEnrichment contains additional data for booking detail
type BookingEnrichment struct {
	ServiceName  string
	CategoryName string
	ChildName    string
	VendorName   string
}
