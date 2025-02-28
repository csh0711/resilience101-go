package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Transaction struct {
	TransactionID uuid.UUID `json:"transactionId,omitempty"` // Auto-generated if missing
	OrderID       uuid.UUID `json:"orderId" validate:"required"`
	UserID        uuid.UUID `json:"userId" validate:"required"`
	Items         []string  `json:"items" validate:"required,dive"`
}

type Result struct {
	TransactionID uuid.UUID `json:"transactionId"`
}

var (
	mightFail bool
	rng       = rand.New(rand.NewSource(time.Now().UnixNano())) // Use a new rand source
)

func LogTransaction(c echo.Context) error {
	failPseudoRandomly()

	var transaction Transaction
	if err := c.Bind(&transaction); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if transaction.TransactionID == uuid.Nil {
		transaction.TransactionID = uuid.New()
	}

	log.Printf("Received transaction: %+v\n", transaction)

	return c.JSON(http.StatusCreated, Result{TransactionID: transaction.TransactionID})
}

func failPseudoRandomly() {
	if mightFail && rng.Intn(2) == 0 { // 50% chance of failure
		log.Println("Failed to create transaction log")
		panic("failed to create transaction log")
	}
}

func main() {
	e := echo.New()

	// Load feature toggle from environment variables
	mightFail = os.Getenv("MIGHT_FAIL") == "true"
	log.Printf("Feature toggle 'mightFail' is set to: %v", mightFail)

	e.POST("/transactions", LogTransaction)

	log.Println("Transaction log service running on port 8082")
	e.Start(":8082")
}
