package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *Server {
	server, err := NewServer()
	require.NoError(t, err)

	// Insert test data
	users := []models.User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
	}

	for _, user := range users {
		err := server.db.Create(&user).Error
		require.NoError(t, err)
	}

	orders := []models.Order{
		{ID: 1, UserID: 1, Amount: 100, CreatedAt: time.Now()},
		{ID: 2, UserID: 1, Amount: 200, CreatedAt: time.Now()},
		{ID: 3, UserID: 2, Amount: 300, CreatedAt: time.Now()},
	}

	for _, order := range orders {
		err := server.db.Create(&order).Error
		require.NoError(t, err)
	}

	// Refresh caches
	require.NoError(t, server.userCache.Refresh())
	require.NoError(t, server.orderCache.Refresh())

	return server
}

func TestHandleGetUser(t *testing.T) {
	server := setupTestServer(t)

	t.Run("get existing user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user?id=1", nil)
		w := httptest.NewRecorder()

		server.handleGetUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user models.User
		err := json.NewDecoder(w.Body).Decode(&user)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", user.Name)
	})

	t.Run("user not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user?id=999", nil)
		w := httptest.NewRecorder()

		server.handleGetUser(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleGetUserOrders(t *testing.T) {
	server := setupTestServer(t)

	t.Run("get user orders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/orders?user_id=1", nil)
		w := httptest.NewRecorder()

		server.handleGetUserOrders(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var orders []models.Order
		err := json.NewDecoder(w.Body).Decode(&orders)
		require.NoError(t, err)
		assert.Len(t, orders, 2)
	})
}

func TestHandleSearchUsers(t *testing.T) {
	server := setupTestServer(t)

	t.Run("search users by name", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/search?name=John", nil)
		w := httptest.NewRecorder()

		server.handleSearchUsers(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var users []models.User
		err := json.NewDecoder(w.Body).Decode(&users)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "John Doe", users[0].Name)
	})
}

func TestHandleHighValueOrders(t *testing.T) {
	server := setupTestServer(t)

	t.Run("get high value orders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/orders/high-value?min_amount=200", nil)
		w := httptest.NewRecorder()

		server.handleHighValueOrders(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var orders []models.Order
		err := json.NewDecoder(w.Body).Decode(&orders)
		require.NoError(t, err)
		assert.Len(t, orders, 2)
		for _, order := range orders {
			assert.GreaterOrEqual(t, order.Amount, float64(200))
		}
	})
}
