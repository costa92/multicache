package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/costa92/multicache/cache"
	"github.com/costa92/multicache/loader"
	"github.com/costa92/multicache/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Server struct {
	db          *gorm.DB
	userCache   *cache.CacheManager[models.UserV2]
	orderCache  *cache.RelatedCacheManager[models.Order]
	userLoader  *loader.GormLoader[models.UserV2]
	orderLoader *loader.GormLoader[models.Order]
}

func NewServer() (*Server, error) {
	// Check if the database file exists and delete it if it does
	if _, err := os.Stat("test.db"); err == nil {
		if err := os.Remove("test.db"); err != nil {
			return nil, fmt.Errorf("failed to delete existing database file: %w", err)
		}
	}

	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate schemas
	if err := db.AutoMigrate(&models.UserV2{}, &models.Order{}); err != nil {
		return nil, err
	}

	// Insert test data using bulk insert for efficiency
	users := []models.UserV2{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
	}

	if err := db.Create(&users).Error; err != nil {
		return nil, err
	}

	orders := []models.Order{
		{ID: 1, UserID: 1, Amount: 100, CreatedAt: time.Now()},
		{ID: 2, UserID: 1, Amount: 200, CreatedAt: time.Now()},
		{ID: 3, UserID: 2, Amount: 300, CreatedAt: time.Now()},
	}

	if err := db.Create(&orders).Error; err != nil {
		return nil, err
	}

	// Initialize loaders
	userLoader := loader.NewGormLoader(db, models.UserV2{}).WithDebug(true)
	orderLoader := loader.NewGormLoader(db, models.Order{}).
		WithJoinsModel(models.UserV2{}, "orders.user_id", "user_v2.id").
		WithDebug(true)

	// Initialize caches with TTL
	userCache := cache.NewCacheManager[models.UserV2](userLoader).WithTTL(5 * time.Minute)
	orderCache := cache.NewRelatedCacheManager[models.Order](orderLoader, 1*time.Minute)

	// Refresh caches
	if err := userCache.Refresh(); err != nil {
		return nil, err
	}
	if err := orderCache.Refresh(); err != nil {
		return nil, err
	}

	return &Server{
		db:          db,
		userCache:   userCache,
		orderCache:  orderCache,
		userLoader:  userLoader,
		orderLoader: orderLoader,
	}, nil
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	user, err := s.userCache.Get(uint(userID))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleGetUserOrders(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("user_id")
	if id == "" {
		http.Error(w, "missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		http.Error(w, "invalid user_id parameter", http.StatusBadRequest)
		return
	}

	orders := s.orderCache.GetByForeignKey(uint(userID))
	json.NewEncoder(w).Encode(orders)
}

func (s *Server) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}

	nameCondition := cache.StringFieldCondition[models.UserV2]{
		FieldExtractor: func(u models.UserV2) string { return u.Name },
		Value:          name,
		Operation:      "contains",
	}

	users := s.userCache.Query(nameCondition)
	json.NewEncoder(w).Encode(users)
}

func (s *Server) handleHighValueOrders(w http.ResponseWriter, r *http.Request) {
	amountStr := r.URL.Query().Get("min_amount")
	if amountStr == "" {
		http.Error(w, "missing min_amount parameter", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "invalid min_amount parameter", http.StatusBadRequest)
		return
	}

	amountCondition := cache.NumberFieldCondition[models.Order, float64]{
		FieldExtractor: func(o models.Order) float64 { return o.Amount },
		Value:          amount,
		Operation:      "gte",
	}

	orders := s.orderCache.Query(amountCondition)
	json.NewEncoder(w).Encode(orders)
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	http.HandleFunc("/api/user", server.handleGetUser)
	http.HandleFunc("/api/user/orders", server.handleGetUserOrders)
	http.HandleFunc("/api/users/search", server.handleSearchUsers)
	http.HandleFunc("/api/orders/high-value", server.handleHighValueOrders)

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
