package jobs

import (
	"github.com/stretchr/testify/mock"
)

// MockAlertService is a mock implementation of AlertService for testing
type MockAlertService struct {
	mock.Mock
}

func (m *MockAlertService) GenerateOverdueAlerts() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAlertService) GenerateReminderAlerts() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAlertService) CleanupResolvedAlerts() error {
	args := m.Called()
	return args.Error(0)
}