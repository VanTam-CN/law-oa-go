package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/tests/test"
)

func TestClientService_CreateClient(t *testing.T) {
	mockClientRepo := new(test.MockClientRepository)
	clientService := NewClientService(mockClientRepo)

	t.Run("Create Client Success", func(t *testing.T) {
		req := &CreateClientRequest{
			Name:    "Test Client",
			Email:   "client@example.com",
			Phone:   "1234567890",
			Address: "123 Test St",
			Company: "Test Company",
			Notes:   "Test client notes",
		}

