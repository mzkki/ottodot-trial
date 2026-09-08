package main

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mzkki/ottodot-trial/internal/domain"
)

func TestLoadTemplates(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	templatesDir := filepath.Join(projectRoot, "templates")

	tmpl, err := loadTemplates(templatesDir)
	if err != nil {
		t.Fatalf("loadTemplates failed: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}

	testCases := []struct {
		name string
		data any
	}{
		{
			name: "home",
			data: map[string]any{
				"Title": "Home",
				"Parents": []domain.Parent{
					{ID: "p1", Name: "Parent 1", Email: "p1@example.com"},
				},
			},
		},
		{
			name: "admin_roster",
			data: map[string]any{
				"Title": "Admin Roster",
				"Classes": []domain.TrialClass{
					{ID: "c1", Title: "Robotics", Instructor: "Mr. X", ScheduledAt: time.Now()},
				},
			},
		},
		{
			name: "student_select",
			data: map[string]any{
				"Students": []domain.Student{
					{ID: "s1", Name: "Student 1", Age: 7},
				},
			},
		},
		{
			name: "class_list",
			data: map[string]any{
				"StudentID": "s1",
				"Classes": []map[string]any{
					{
						"ID":              "c1",
						"Title":           "Robotics",
						"SpotsLeft":       2,
						"Description":     "Fun class",
						"Instructor":      "Mr. X",
						"ScheduledAt":     "Mon, 01 Jan",
						"DurationMinutes": 60,
						"IsBooked":        false,
					},
					{
						"ID":              "c2",
						"Title":           "Coding",
						"SpotsLeft":       1,
						"Description":     "Fun coding",
						"Instructor":      "Ms. Y",
						"ScheduledAt":     "Tue, 02 Jan",
						"DurationMinutes": 60,
						"IsBooked":        true,
						"BookingStatus":   "confirmed",
						"BookingID":       "b2",
					},
					{
						"ID":              "c3",
						"Title":           "Art",
						"SpotsLeft":       1,
						"Description":     "Fun art",
						"Instructor":      "Mrs. Z",
						"ScheduledAt":     "Wed, 03 Jan",
						"DurationMinutes": 45,
						"IsBooked":        true,
						"BookingStatus":   "pending",
						"BookingID":       "b3",
					},
				},
			},
		},
		{
			name: "booking_form",
			data: map[string]any{
				"Booking": domain.Booking{
					ID:     "b1",
					Status: domain.BookingStatusPending,
					Student: domain.Student{
						Name: "Emma",
					},
					TrialClass: domain.TrialClass{
						Title:      "Robotics",
						Instructor: "Mr. Smith",
					},
				},
			},
		},
		{
			name: "booking_status",
			data: map[string]any{
				"Booking": domain.Booking{
					ID:     "b1",
					Status: domain.BookingStatusConfirmed,
					Student: domain.Student{
						Name: "Emma",
					},
					TrialClass: domain.TrialClass{
						Title:       "Robotics",
						Instructor:  "Mr. Smith",
						ScheduledAt: time.Now(),
					},
					PaymentAttempts: []domain.PaymentAttempt{
						{
							Status:      "success",
							AmountCents: 50000,
							CreatedAt:   time.Now(),
						},
					},
				},
				"ErrorMessage": "",
			},
		},
		{
			name: "roster_table",
			data: map[string]any{
				"Class": domain.TrialClass{
					Title:       "Robotics",
					Instructor:  "Mr. Smith",
					ScheduledAt: time.Now(),
					MaxCapacity: 4,
				},
				"Bookings": []domain.Booking{
					{
						ID:        "b1",
						CreatedAt: time.Now(),
						Student: domain.Student{
							Name: "Emma",
							Age:  7,
							Parent: domain.Parent{
								Name:  "Sarah",
								Email: "sarah@example.com",
							},
						},
					},
				},
				"Count": 1,
			},
		},
		{
			name: "error",
			data: map[string]any{
				"Title":   "Error",
				"Message": "Something broke",
			},
		},
		{
			name: "error_partial",
			data: map[string]any{
				"Message": "Something broke partial",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, tc.name, tc.data); err != nil {
				t.Fatalf("failed to execute template %s: %v", tc.name, err)
			}
			if buf.Len() == 0 {
				t.Fatalf("template %s produced empty output", tc.name)
			}
		})
	}
}

func TestAppRouterEndpoints(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	_ = os.Chdir(projectRoot)

	_ = os.Setenv("APP_ENV", "test")
	app := newApp()
	if app == nil || app.Router == nil {
		t.Fatal("expected non-nil app and router")
	}

	endpoints := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/api/health", 200},
		{"GET", "/", 200},
		{"GET", "/admin/roster", 200},
		{"GET", "/api/v1/classes/all", 200},
		{"GET", "/api/v1/students", 200},
		{"GET", "/api/v1/classes", 200},
		{"GET", "/api/v1/roster", 200},
		{"POST", "/api/v1/bookings", 200}, // validation error partial
		{"GET", "/docs", 301},
		{"GET", "/api/v1/docs", 200},
		{"GET", "/api/v1/docs/", 200},
		{"GET", "/api/v1/docs/openapi.json", 200},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(ep.method, ep.path, nil)
			app.Router.ServeHTTP(w, req)

			if w.Code != ep.code {
				t.Errorf("expected status %d for %s %s, got %d", ep.code, ep.method, ep.path, w.Code)
			}
		})
	}
}
