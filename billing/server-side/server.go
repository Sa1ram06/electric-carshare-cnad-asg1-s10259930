package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Card struct
type Card struct {
	CardID      int     `json:"card_id"`
	CardNumber  string  `json:"card_number"`
	CardExpiry  string  `json:"card_expiry"`
	CVV         string  `json:"cvv"`
	CardBalance float64 `json:"card_balance"`
	UserID      int     `json:"user_id"`
}

// Invoice struct
type Invoice struct {
	InvoiceID       int     `json:"invoice_id"`
	BookingID       int     `json:"booking_id"`
	UserID          int     `json:"user_id"`
	IssueDate       string  `json:"issue_date"`
	BaseCost        float64 `json:"base_cost"`
	PromotionCode   *string `json:"promo_code"`
	DiscountApplied float64 `json:"discount_applied"`
	TotalAmount     float64 `json:"total_amount"`
	Details         string  `json:"details"`
	Status          string  `json:"status"`
}

type Billing struct {
	BillingID         int     `json:"billing_id"`
	InvoiceID         int     `json:"invoice_id"`
	CardID            int     `json:"card_id"`
	TransactionAmount float64 `json:"transaction_amount"`
	TransactionDate   string  `json:"transaction_date"`
}

type VehicleBookingDetails struct {
	BookingID          int64   `json:"booking_id"`
	ScheduleID         int64   `json:"schedule_id"`
	UserID             int     `json:"user_id"`
	Status             string  `json:"status"`
	BaseCost           float64 `json:"base_cost"`
	PromotionCode      *string `json:"promo_code"`
	MembershipDiscount float64 `json:"membership_discount"`
	PromotionDiscount  float64 `json:"promotion_discount"`
	DiscountApplied    float64 `json:"discount_applied"`
	TotalAmount        float64 `json:"total_amount"`
	Type               string  `json:"type"`
	Brand              string  `json:"brand"`
	Model              string  `json:"model"`
	LicensePlate       string  `json:"license_plate"`
	ScheduleDate       string  `json:"date"`
	StartTime          string  `json:"start_time"`
	EndTime            string  `json:"end_time"`
}

// Receipt struct
type Receipt struct {
	ReceiptID     int     `json:"receipt_id"`
	BillingID     int     `json:"billing_id"`
	CardID        int     `json:"card_id"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
	Description   string  `json:"description"`
	CardLastThree string  `json:"card_last_three"`
}

var db *sql.DB

// Initialise the user_svc_db database connection
func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/billing_svc_db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
		return
	}
}

func main() {
	// Call initDB(), to initialise user_svc_db connection
	initDB()
	defer db.Close()
	// Setting up router and API endpoints
	router := mux.NewRouter()
	handler := cors.Default().Handler(router)
	router.HandleFunc("/api/v1/card-details/{id}", getCardDetailsByUserID).Methods("GET")
	router.HandleFunc("/api/v1/create-invoice/{id}/{booking_id}", createInvoice).Methods("POST")
	router.HandleFunc("/api/v1/invoice-details/{id}", getInvoiceDetailsByUserID).Methods("GET")
	router.HandleFunc("/api/v1/invoice-details-by-id/{id}", getInvoiceDetailsByInvoiceID).Methods("GET")
	router.HandleFunc("/api/v1/make-payment/{id}", makePayment).Methods("POST")
	router.HandleFunc("/api/v1/receipt-details/{id}", getReceiptDetailsByBillingID).Methods("GET")
	fmt.Println("Listening at port 8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}

// Validate Booking
func validateBooking(UserId string, BookingId string) (*VehicleBookingDetails, error) {
	// Struct for response from the booking service
	type Response struct {
		Message string                `json:"message"`
		Booking VehicleBookingDetails `json:"booking"`
	}

	// URL of the booking service
	bookingServiceURL := "http://localhost:9000/api/v1/verify-booking/" + UserId + "/" + BookingId

	// Send GET request to the booking service to validate the booking
	resp, err := http.Get(bookingServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking data: %v", err)
	}
	defer resp.Body.Close()

	// Create a Response object to hold the data returned from the booking service
	var response Response

	switch resp.StatusCode {
	case http.StatusOK:
		// Decode the response into the Response struct
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode booking data: %v", err)
		}
		return &response.Booking, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("booking not found")

	default:
		return nil, fmt.Errorf("failed to get booking data, status code: %d", resp.StatusCode)
	}
}

// Get Card Details by User ID
func getCardDetailsByUserID(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string `json:"message"`
		Card    *Card  `json:"card"`
	}

	// Get the user_id from the request
	userId := mux.Vars(r)["id"]

	// Query to get card details by user_id
	query := "SELECT card_id, card_number, card_expiry, cvv, card_balance FROM card WHERE user_id = ?"

	// Execute the query
	row := db.QueryRow(query, userId)

	// Scan the row and assign to the Card struct
	var card Card
	if err := row.Scan(&card.CardID, &card.CardNumber, &card.CardExpiry, &card.CVV, &card.CardBalance); err != nil {
		// If there is an error
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Card not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		// If there is an error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying card", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error converting user_id to int", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	card.UserID = userIdInt
	// If card found
	w.WriteHeader(http.StatusOK)
	response := Response{"Card found", &card}
	json.NewEncoder(w).Encode(response)
}

// Create Invoice
func createInvoice(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string   `json:"message"`
		Invoice *Invoice `json:"invoice"`
	}
	// Get the user_id and booking_id from the request
	userId := mux.Vars(r)["id"]
	bookingId := mux.Vars(r)["booking_id"]

	// Validate the booking
	booking, err := validateBooking(userId, bookingId)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Booking not found", nil} // comment
		json.NewEncoder(w).Encode(response)
		return
	}
	fmt.Println(booking)
	baseCost := booking.BaseCost
	promotionCode := booking.PromotionCode
	fmt.Println(promotionCode)
	discountApplied := booking.DiscountApplied
	totalAmount := booking.TotalAmount
	details := "Reserved the " + booking.Brand + " " + booking.Model + " on " + booking.ScheduleDate + " from " + booking.StartTime + " to " + booking.EndTime
	var invoiceSentBefore bool
	var invoiceId int64
	// Query to check if invoice is already sent
	query := "SELECT invoice_id FROM invoice WHERE booking_id = ?"
	err = db.QueryRow(query, bookingId).Scan(&invoiceId) // corrected the assignment here
	if err != nil {
		if err == sql.ErrNoRows {
			invoiceSentBefore = false
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Error querying invoice", nil} // comment
			json.NewEncoder(w).Encode(response)
			return
		}
	} else {
		invoiceSentBefore = true
	}
	// If invoice is not sent before
	if !invoiceSentBefore {
		if promotionCode != nil {
			// Insert query with promo code
			query = "INSERT INTO invoice (booking_id, user_id, base_cost, promo_code, discount_applied, total_amount, details, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
			result, err := db.Exec(query, bookingId, userId, baseCost, promotionCode, discountApplied, totalAmount, details, "Pending")
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				response := Response{"Error inserting invoice", nil}
				json.NewEncoder(w).Encode(response)
				return
			}
			invoiceId, err = result.LastInsertId()
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				response := Response{"Error getting invoice id", nil}
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			// Insert query without promo code
			query = "INSERT INTO invoice (booking_id, user_id, base_cost, discount_applied, total_amount, details, status) VALUES (?, ?, ?, ?, ?, ?, ?)"
			result, err := db.Exec(query, bookingId, userId, baseCost, discountApplied, totalAmount, details, "Pending")
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				response := Response{"Error inserting invoice", nil}
				json.NewEncoder(w).Encode(response)
				return
			}
			invoiceId, err = result.LastInsertId()
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				response := Response{"Error getting invoice id", nil}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		var invoice Invoice
		// Get the invoice details
		if promotionCode != nil { // corrected the condition to check promo code
			query = "SELECT invoice_id, booking_id, user_id, issue_date, base_cost, promo_code, discount_applied, total_amount, details, status FROM invoice WHERE invoice_id = ?"
			err = db.QueryRow(query, invoiceId).Scan(&invoice.InvoiceID, &invoice.BookingID, &invoice.UserID, &invoice.IssueDate, &invoice.BaseCost, &invoice.PromotionCode, &invoice.DiscountApplied, &invoice.TotalAmount, &invoice.Details, &invoice.Status)
		} else {
			query = "SELECT invoice_id, booking_id, user_id, issue_date, base_cost, discount_applied, total_amount, details, status FROM invoice WHERE invoice_id = ?"
			err = db.QueryRow(query, invoiceId).Scan(&invoice.InvoiceID, &invoice.BookingID, &invoice.UserID, &invoice.IssueDate, &invoice.BaseCost, &invoice.DiscountApplied, &invoice.TotalAmount, &invoice.Details, &invoice.Status)
		}
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Error querying invoice", nil} // comment
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := Response{"Invoice created", &invoice}
		json.NewEncoder(w).Encode(response)
	}
	// If invoice is already sent
	if invoiceSentBefore {
		w.WriteHeader(http.StatusConflict)
		response := Response{"Invoice already sent", nil}
		json.NewEncoder(w).Encode(response)
	}
}

// Get multiple/1/none invoice details by user_id
func getInvoiceDetailsByUserID(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message  string    `json:"message"`
		Invoices []Invoice `json:"invoices"`
	}

	// Get the user_id from the request
	userId := mux.Vars(r)["id"]

	// Query to get invoice details by user_id
	query := `SELECT invoice_id, booking_id, user_id, issue_date, base_cost, promo_code, discount_applied, total_amount, details, status 
	FROM 
		invoice 
	WHERE 
		user_id = ? 
	ORDER BY 
		CASE 
			WHEN status = 'Pending' THEN 1 
			WHEN status = 'Paid' THEN 2 
			ELSE 3 
		END, 
		issue_date DESC;`
	// Execute the query
	rows, err := db.Query(query, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying invoices", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	// Slice to hold the invoices
	var invoices []Invoice

	// Iterate over the rows
	for rows.Next() {
		var invoice Invoice
		// Scan the row and assign to the Invoice struct
		if err := rows.Scan(&invoice.InvoiceID, &invoice.BookingID, &invoice.UserID, &invoice.IssueDate, &invoice.BaseCost, &invoice.PromotionCode, &invoice.DiscountApplied, &invoice.TotalAmount, &invoice.Details, &invoice.Status); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Error iterating invoices", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		// Append the invoice to the invoices slice
		invoices = append(invoices, invoice)
	}

	// Check for any error encountered during the rows iteration
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error iterating invoices", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// If invoices found
	if len(invoices) > 0 {
		w.WriteHeader(http.StatusOK)
		response := Response{"Invoices found", invoices}
		json.NewEncoder(w).Encode(response)
	} else {
		// If no invoices found
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No invoices found", nil}
		json.NewEncoder(w).Encode(response)
	}
}

// Get ibvoice details by invoice_id
func getInvoiceDetailsByInvoiceID(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string   `json:"message"`
		Invoice *Invoice `json:"invoice"`
	}

	// Get the invoice_id from the request
	invoiceId := mux.Vars(r)["id"]

	// Query to get invoice details by invoice_id
	query := `SELECT invoice_id, booking_id, user_id, issue_date, base_cost, promo_code, discount_applied, total_amount, details, status 
	FROM 
		invoice 
	WHERE 
		invoice_id = ?;`

	// Execute the query
	var invoice Invoice
	err := db.QueryRow(query, invoiceId).Scan(&invoice.InvoiceID, &invoice.BookingID, &invoice.UserID, &invoice.IssueDate, &invoice.BaseCost, &invoice.PromotionCode, &invoice.DiscountApplied, &invoice.TotalAmount, &invoice.Details, &invoice.Status)
	if err != nil {
		// If there is an error
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Invoice not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		// If there is another error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying invoice", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// If invoice found
	w.WriteHeader(http.StatusOK)
	response := Response{"Invoice found", &invoice}
	json.NewEncoder(w).Encode(response)
}

// Make Payment
func makePayment(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string   `json:"message"`
		Billing *Billing `json:"billing"`
	}

	// Get the invoice_id from the request
	invoiceId := mux.Vars(r)["id"]

	// Query to get invoice details by invoice_id
	query := "SELECT booking_id, user_id, total_amount, status  FROM invoice WHERE invoice_id = ?"

	var bookingId int
	var userId int
	var totalAmount float64
	var status string
	// Execute the query
	err := db.QueryRow(query, invoiceId).Scan(&bookingId, &userId, &totalAmount, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Invoice not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying invoice", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	if status == "Paid" {
		w.WriteHeader(http.StatusConflict)
		response := Response{"Invoice already paid", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Get the card details from the request
	var card Card
	err = json.NewDecoder(r.Body).Decode(&card)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Invalid card details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get card details by user_id
	var cardDetails Card
	query = "SELECT card_id, card_number, card_expiry, cvv, card_balance FROM card WHERE user_id = ?"
	err = db.QueryRow(query, userId).Scan(&cardDetails.CardID, &cardDetails.CardNumber, &cardDetails.CardExpiry, &cardDetails.CVV, &cardDetails.CardBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Card not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying card details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check card number matches with the card details
	if card.CardNumber != cardDetails.CardNumber {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Card number does not match", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check card expiry matches with the card details
	if card.CardExpiry != cardDetails.CardExpiry {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Card expiry does not match", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check CVV matches with the card details
	if card.CVV != cardDetails.CVV {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"CVV does not match", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check the expiry date of the card
	today := time.Now()
	expiryDate, err := time.Parse("01/06", cardDetails.CardExpiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error parsing expiry date", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	if today.After(expiryDate) {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Card expired", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check if the card has sufficient balance
	if cardDetails.CardBalance < totalAmount {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Insufficient balance", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Update the card balance
	newBalance := cardDetails.CardBalance - totalAmount
	query = "UPDATE card SET card_balance = ? WHERE user_id = ?"
	_, err = db.Exec(query, newBalance, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error updating card balance", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Insert the billing details
	transactionDate := time.Now().Format("2006-01-02")
	query = "INSERT INTO billing (invoice_id, card_id, transaction_amount, transaction_date) VALUES (?, ?, ?, ?)"
	result, err := db.Exec(query, invoiceId, cardDetails.CardID, totalAmount, transactionDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error inserting billing details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Make payment and then confirm booking
	fmt.Println("userId: ", userId, "bookingId: ", bookingId)
	bookingConfirmationURL := "http://localhost:9000/api/v1/confirm-booking/" + strconv.Itoa(userId) + "/" + strconv.Itoa(bookingId)
	fmt.Println("bookingConfirmationURL: ", bookingConfirmationURL)
	var paymentConfirmation = struct {
		Message        string `json:"message"`
		PaymentSuccess bool   `json:"paymentSuccess"`
	}{
		Message:        "Payment successful",
		PaymentSuccess: true, // Assume payment is successful, set to true
	}

	// Prepare JSON payload for booking confirmation
	jsonData, err := json.Marshal(paymentConfirmation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error preparing confirmation data", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send the confirmation request
	resp, err := http.Post(bookingConfirmationURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error sending booking confirmation", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	fmt.Println("Booking confirmation response status: ", resp.Status)
	defer resp.Body.Close()

	// Get the billing id
	billingId, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error getting billing id", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	var billing Billing
	// Get the billing details
	query = "SELECT billing_id, invoice_id, card_id, transaction_amount, transaction_date FROM billing WHERE billing_id = ?"
	err = db.QueryRow(query, billingId).Scan(&billing.BillingID, &billing.InvoiceID, &billing.CardID, &billing.TransactionAmount, &billing.TransactionDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying billing details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{"Payment successful", &billing}
	json.NewEncoder(w).Encode(response)
}

// Get Receipt Details by Billing ID
func getReceiptDetailsByBillingID(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string   `json:"message"`
		Receipt *Receipt `json:"receipt"`
	}

	// Get the billing_id from the request
	billingId := mux.Vars(r)["id"]

	// Query to get receipt details and the card number
	query := `
		SELECT r.receipt_id, r.card_id, r.amount, r.date, r.description, c.card_number
		FROM receipt r
		INNER JOIN card c ON r.card_id = c.card_id
		WHERE r.billing_id = ?
	`

	// Declare variables to hold the receipt data and card number
	var receipt Receipt
	var cardNumber string

	// Execute the query
	err := db.QueryRow(query, billingId).Scan(&receipt.ReceiptID, &receipt.CardID, &receipt.Amount, &receipt.Date, &receipt.Description, &cardNumber)
	if err != nil {
		// If there is an error
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Receipt not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		// If there is another error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error querying receipt", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	receipt.BillingID, _ = strconv.Atoi(billingId)
	// Mask the card number by replacing the first digits with asterisks and keeping the last 3 digits
	maskedCardNumber := maskCardNumber(cardNumber)

	// Assign the masked card number to the receipt
	receipt.CardLastThree = maskedCardNumber

	// If receipt found
	w.WriteHeader(http.StatusOK)
	response := Response{"Receipt found", &receipt}
	json.NewEncoder(w).Encode(response)
}

// Mask the card number by replacing all but the last 3 digits with asterisks
func maskCardNumber(cardNumber string) string {
	// Ensure the card number is at least 3 characters long
	if len(cardNumber) < 3 {
		return cardNumber
	}
	// Mask the card number (e.g., ************123)
	return "**** **** **** " + cardNumber[len(cardNumber)-4:]
}
