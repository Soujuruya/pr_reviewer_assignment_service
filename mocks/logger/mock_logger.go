package mocks

import (
	"context"
	"pr_reviewer_assignment_service/pkg/logger"

	"go.uber.org/zap"
)

type MockLogger struct{}

func NewMockLogger() logger.Logger {
	return &MockLogger{}
}

func (l *MockLogger) Info(ctx context.Context, msg string, fields ...zap.Field)  {}
func (l *MockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {}
func (l *MockLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {}
func (l *MockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field)  {}
func (l *MockLogger) Sync()                                                      {}
