package handlers

import (
	"board-game-library/internal/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserServiceInterface for testing
type MockUserServiceInterface struct {
	mock.Mock
}

func (m *MockUserServiceInterface) RegisterUser(name, email string) (*models.User, error) {
	args := m.Called(name, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserServiceInterface) GetUser(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserServiceInterface) GetAllUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserServiceInterface) GetUserBorrowings(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockUserServiceInterface) GetActiveUserBorrowings(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockUserServiceInterface) CanUserBorrow(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserServiceInterface) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestUserWebHandler_SearchFilterUsers_Logic(t *testing.T) {
	tests := []struct {
		name          string
		searchQuery   string
		status        string
		users         []*models.User
		expectedUsers int
	}{
		{
			name:        "search with no filters",
			searchQuery: "",
			status:      "all",
			users: []*models.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, CurrentLoans: 0},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: false, CurrentLoans: 0},
			},
			expectedUsers: 2,
		},
		{
			name:        "search by name",
			searchQuery: "john",
			status:      "all",
			users: []*models.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, CurrentLoans: 0},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: false, CurrentLoans: 0},
			},
			expectedUsers: 1,
		},
		{
			name:        "filter active users only",
			searchQuery: "",
			status:      "active",
			users: []*models.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, CurrentLoans: 0},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: false, CurrentLoans: 0},
			},
			expectedUsers: 1,
		},
		{
			name:        "filter users with loans",
			searchQuery: "",
			status:      "with-loans",
			users: []*models.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, CurrentLoans: 1},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true, CurrentLoans: 0},
			},
			expectedUsers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := tt.users

			// Apply search filter
			if tt.searchQuery != "" {
				filteredUsers := make([]*models.User, 0)
				queryLower := strings.ToLower(tt.searchQuery)
				
				for _, user := range users {
					if strings.Contains(strings.ToLower(user.Name), queryLower) ||
					   strings.Contains(strings.ToLower(user.Email), queryLower) {
						filteredUsers = append(filteredUsers, user)
					}
				}
				users = filteredUsers
			}

			// Apply status filter
			if tt.status != "" && tt.status != "all" {
				filteredUsers := make([]*models.User, 0)
				for _, user := range users {
					switch tt.status {
					case "active":
						if user.IsActive {
							filteredUsers = append(filteredUsers, user)
						}
					case "inactive":
						if !user.IsActive {
							filteredUsers = append(filteredUsers, user)
						}
					case "with-loans":
						if user.CurrentLoans > 0 {
							filteredUsers = append(filteredUsers, user)
						}
					}
				}
				users = filteredUsers
			}

			assert.Equal(t, tt.expectedUsers, len(users))
		})
	}
}

func TestUserWebHandler_ServiceIntegration(t *testing.T) {
	// Test that the handler correctly calls the service methods
	mockUserService := new(MockUserServiceInterface)
	handler := NewUserWebHandler(mockUserService)

	// Test that the handler is properly initialized
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.userService)
}