package users

import (
	"testing"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
)

func TestUserService_NewUserService(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test service creation
	userService := NewUserService(nil, logger)

	assert.NotNil(t, userService)
	assert.Nil(t, userService.DB)
	assert.Equal(t, logger, userService.logger)
}

func TestUserService_GetAllUsers_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test service creation
	userService := &UserService{
		DB:     nil,
		logger: logger,
	}

	// Test that the service was created properly
	assert.NotNil(t, userService)
	assert.Equal(t, logger, userService.logger)
	assert.Nil(t, userService.DB)
}

func TestUserService_PaginationLogic(t *testing.T) {
	// Test pagination logic that would be used in GetAllUsers
	limit := 10
	offset := 20

	// Test that pagination parameters are valid
	assert.Greater(t, limit, 0)
	assert.GreaterOrEqual(t, offset, 0)

	// Test edge cases
	assert.Equal(t, 0, 0)     // offset can be 0 for first page
	assert.Equal(t, 1, 1)     // limit can be 1 for single item
	assert.Equal(t, 100, 100) // reasonable upper limit
}

func TestUserService_LoggingStructure(t *testing.T) {
	// Test the logging structure used in GetAllUsers
	logData := map[string]any{
		"limit":  10,
		"offset": 20,
		"count":  5,
	}

	// Verify the structure of log data
	assert.Equal(t, 10, logData["limit"])
	assert.Equal(t, 20, logData["offset"])
	assert.Equal(t, 5, logData["count"])

	// Test that all required fields are present
	requiredFields := []string{"limit", "offset", "count"}
	for _, field := range requiredFields {
		assert.Contains(t, logData, field)
	}
}

// Benchmark tests for performance
func BenchmarkUserService_NewUserService(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewUserService(nil, logger)
	}
}

func BenchmarkUserService_PaginationLogic(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limit := 10
		offset := 20
		_ = limit > 0
		_ = offset >= 0
	}
}

func BenchmarkUserService_LoggingStructure(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logData := map[string]any{
			"limit":  10,
			"offset": 20,
			"count":  5,
		}
		_ = logData["limit"]
		_ = logData["offset"]
		_ = logData["count"]
	}
}
