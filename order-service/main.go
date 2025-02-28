package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Order represents an order request
type Order struct {
	OrderID uuid.UUID   `json:"orderId"` // Nullable in Kotlin, optional in Go
	UserID  uuid.UUID   `json:"userId" validate:"required"`
	Items   []OrderItem `json:"items" validate:"required,dive"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ItemID   uuid.UUID `json:"itemId" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,gt=0"` // Quantity must be > 0
}

// Transaction represents a transaction log entry sent to transaction-log-service
type Transaction struct {
	TransactionID uuid.UUID `json:"transactionId,omitempty"` // Auto-generated if missing
	OrderID       uuid.UUID `json:"orderId"`
	UserID        uuid.UUID `json:"userId"`
	Items         []string  `json:"items"`
}

// CreateOrder handles the POST /orders request
func CreateOrder(c echo.Context) error {
	var order Order
	if err := c.Bind(&order); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// If OrderID is empty, generate a new UUID
	if order.OrderID == uuid.Nil {
		order.OrderID = uuid.New()
	}

	log.Printf("Received order: %+v\n", order)

	// Send transaction log request
	if err := logTransaction(order); err != nil {
		log.Println("Failed to log transaction:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to log transaction"})
	}

	return c.JSON(http.StatusCreated, order)
}

// logTransaction sends a transaction log request to transaction-log-service
func logTransaction(order Order) error {
	// Convert OrderItem list to a string list for transaction logging
	var itemNames []string
	for _, item := range order.Items {
		itemNames = append(itemNames, item.Name)
	}

	// Create Transaction struct
	transaction := Transaction{
		OrderID: order.OrderID,
		UserID:  order.UserID,
		Items:   itemNames,
	}

	jsonData, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8082/transactions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	log.Println("Transaction successfully logged")
	return nil
}

func main() {
	e := echo.New()
	e.POST("/orders", CreateOrder)

	log.Println("Order service running on port 8081")
	e.Start(":8081")
}
