package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Struct to represent the vehicle schedule data
type VehicleSchedules struct {
	ScheduleID   int     `json:"schedule_id"`
	VehicleID    string  `json:"vehicle_id"`
	Type         string  `json:"type"`
	Brand        string  `json:"brand"`
	Model        string  `json:"model"`
	LicensePlate string  `json:"license_plate"`
	HourlyRate   float64 `json:"hourly_rate"`
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	BaseCost     float64 `json:"base_cost"`
}

// Struct to represent the vehicle booking details
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
	HourlyRate         float64 `json:"hourly_rate"`
}

// User Struct (req body)
type User struct {
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	Dob              string `json:"dob"`
	Password         string `json:"password"`
	MembershipId     string `json:"membership_id"`
	LicenseNumber    string `json:"license_number"`
	LicenseExpiry    string `json:"license_expiry"`
	VerificationCode string `json:"verification_code"`
	Verified         bool   `json:"verified"`
}

// Membership struct (req body)
type Membership struct {
	MembershipId       string  `json:"membership_id"`
	HourlyRateDiscount float64 `json:"hourly_rate_discount"`
	BookingLimit       int     `json:"booking_limit"`
}

// Struct to represent promotion details
type Promotion struct {
	PromoCode         string  `json:"promo_code"`
	PromotionName     string  `json:"promotion_name"`
	PromotionDiscount float64 `json:"discount_percentage"`
	ValidFrom         string  `json:"valid_from"`
	ValidTo           string  `json:"valid_to"`
}

var db *sql.DB

// Initialise the user_svc_db database connection
func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/vehicle_svc_db")
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
	router.HandleFunc("/api/v1/vehicles/{date}", getVehicles).Methods("GET")
	router.HandleFunc("/api/v1/vehicle/{scheduleId}", getVehicleDetails).Methods("GET")
	router.HandleFunc("/api/v1/rental-history/{id}", getRentalHistory).Methods("GET")
	router.HandleFunc("/api/v1/upcoming-rentals/{id}", getUpcomingRental).Methods("GET")
	router.HandleFunc("/api/v1/create-booking-session/{id}/{scheduleId}", createBookingSession).Methods("POST")
	router.HandleFunc("/api/v1/add-promotion-code/{id}/{bookingId}/{promoCode}", addPromotionCode).Methods("POST")
	router.HandleFunc("/api/v1/cancel-booking-session/{id}/{bookingId}", deleteBookingSession).Methods("DELETE")
	router.HandleFunc("/api/v1/cancel-booking/{id}/{bookingId}", deleteBooking).Methods("DELETE")
	router.HandleFunc("/api/v1/verify-booking/{id}/{bookingId}", verifyBooking).Methods("GET")
	router.HandleFunc("/api/v1/confirm-booking/{id}/{bookingId}", confirmBooking).Methods("POST")
	router.HandleFunc("/api/v1/vehicle-by-hourly-rate/{hourlyRate}", getVehicleDetailsByHourlyRate).Methods("GET")
	router.HandleFunc("/api/v1/update-booking/{id}/{bookingId}/{scheduleId}", updateBooking).Methods("PUT")
	fmt.Println("Listening at port 9000")
	log.Fatal(http.ListenAndServe(":9000", handler))
}

// Validate date
func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// Validate User
func validateUser(userId string) (*User, error) {
	// Struct for response from the user service
	type Response struct {
		Message string `json:"message"`
		User    User   `json:"user"`
	}

	// URL of the user service
	userServiceURL := "http://localhost:8000/api/v1/validate-user/" + userId

	// Send GET request to the user service to validate the user
	resp, err := http.Get(userServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %v", err)
	}
	defer resp.Body.Close()

	// Create a Response object to hold the data returned from the user service
	var response Response

	switch resp.StatusCode {
	case http.StatusOK:
		// Decode the response into the Response struct
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode user data: %v", err)
		}
		return &response.User, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("user not found")

	default:
		return nil, fmt.Errorf("failed to get user data, status code: %d", resp.StatusCode)
	}
}

// Get membership details
func getMembershipDetails(membershipId string) (*Membership, error) {
	// Struct for response from the user service
	type Response struct {
		Message    string     `json:"message"`
		Membership Membership `json:"membership"`
	}

	// URL of the user service
	membershipServiceURL := "http://localhost:8000/api/v1/membership/" + membershipId

	// Send GET request to the user service to validate the user
	resp, err := http.Get(membershipServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership data: %v", err)
	}
	defer resp.Body.Close()

	// Create a Response object to hold the data returned from the user service
	var response Response

	switch resp.StatusCode {
	case http.StatusOK:
		// Decode the response into the Response struct
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode membership data: %v", err)
		}
		return &response.Membership, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("membership not found")

	default:
		return nil, fmt.Errorf("failed to get membership data, status code: %d", resp.StatusCode)
	}
}

// Get promotion details by promotion code
func getPromotionByPromoCode(promocode string) (*Promotion, error) {
	// Struct for response
	type Response struct {
		Message   string     `json:"message"`
		Promotion *Promotion `json:"promotion"`
	}

	// URL of the promotion service
	promotionServiceURL := "http://localhost:8080/api/v1/promotions/" + promocode

	// Send GET request to the promotion service
	resp, err := http.Get(promotionServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get promotion data: %v", err)
	}
	defer resp.Body.Close()

	// Create a Response object to hold the data returned from the promotion service
	var response Response

	switch resp.StatusCode {
	case http.StatusOK:
		// Decode the response into the Response struct
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode promotion data: %v", err)
		}
		return response.Promotion, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("promotion not found")

	default:
		return nil, fmt.Errorf("failed to get promotion data, status code: %d", resp.StatusCode)
	}
}

// Calculate the total cost of the booking
func calculateAmount(vehicleHourlyRate float64, startTime, endTime time.Time, membershipDiscount float64, promoCode string) (float64, float64, float64, float64, float64, error) {
	// Calculate the duration in hours
	duration := endTime.Sub(startTime).Hours()

	// Calculate the base amount
	baseAmount := vehicleHourlyRate * duration

	// Apply membership discount
	membershipDiscountAmount := baseAmount * (membershipDiscount / 100)
	discountedAmountAfterMembership := baseAmount - membershipDiscountAmount

	// Initialize total discount and total amount after membership discount
	totalDiscount := membershipDiscountAmount
	totalAmount := discountedAmountAfterMembership
	// Apply promo code discount
	if promoCode != "" {
		// Apply promo discount on the amount after membership discount
		promotion, err := getPromotionByPromoCode(promoCode)
		if err != nil {
			return 0, 0, 0, 0, 0, fmt.Errorf("promo code not found")
		}
		// Compare the current date with the promotion valid from and valid to dates
		currentDate := time.Now().Format("2006-01-02")
		if currentDate < promotion.ValidFrom || currentDate > promotion.ValidTo {
			return 0, 0, 0, 0, 0, fmt.Errorf("promo code not valid")
		}
		promotionDiscount := promotion.PromotionDiscount
		promotionDiscountAmount := discountedAmountAfterMembership * (promotionDiscount / 100)
		totalDiscount += promotionDiscountAmount
		totalAmount -= promotionDiscountAmount
		return baseAmount, membershipDiscountAmount, promotionDiscountAmount, totalDiscount, totalAmount, nil
	}

	// Return base amount, membership discount, promo discount, total discount, and final total amount
	return baseAmount, membershipDiscountAmount, 0, totalDiscount, totalAmount, nil
}

// Get all vehicles that has not been reserved and from given the date
func getVehicles(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Variable to hold the vehicles
	var vehicles []VehicleSchedules
	// Struct for response
	type Response struct {
		Message  string             `json:"message"`
		Vehicles []VehicleSchedules `json:"vehicles"`
	}
	date := mux.Vars(r)["date"]
	// Check if the date is in the correct format
	if !isValidDate(date) {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Invalid date format", vehicles}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Query to get all vehicles that has not been reserved and from given the date
	query := `
		SELECT s.schedule_id, v.vehicle_id, v.type, v.brand, v.model, v.license_plate, v.hourly_rate, s.date, s.start_time, s.end_time
		FROM vehicles v
		INNER JOIN schedules s ON v.vehicle_id = s.vehicle_id
		WHERE s.date = ? AND s.is_reserved = 0;
	`
	rows, err := db.Query(query, date)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var vehicle VehicleSchedules
		err := rows.Scan(&vehicle.ScheduleID, &vehicle.VehicleID, &vehicle.Type, &vehicle.Brand, &vehicle.Model, &vehicle.LicensePlate, &vehicle.HourlyRate, &vehicle.Date, &vehicle.StartTime, &vehicle.EndTime)
		if err != nil {
			http.Error(w, "Error reading vehicle data", http.StatusInternalServerError)
			log.Println("Error reading vehicle data:", err)
			return
		}
		vehicles = append(vehicles, vehicle)
	}
	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		log.Println("Error iterating over rows:", err)
		return
	}
	// If no vehicles were found, return a "Not Found" message
	if len(vehicles) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No vehicle available in the given range", vehicles}
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{"Vehicles available in the given range", vehicles}
	json.NewEncoder(w).Encode(response)
}

// Get selected vehicle details
func getVehicleDetails(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Variable to hold the vehicle details
	var vehicle VehicleSchedules
	// Get the schedule_id from the URL
	scheduleId := mux.Vars(r)["scheduleId"]
	// Struct for response
	type Response struct {
		Message string            `json:"message"`
		Vehicle *VehicleSchedules `json:"vehicle"`
	}
	// Query to get the vehicle details
	query := `
		SELECT s.schedule_id, v.vehicle_id, v.type, v.brand, v.model, v.license_plate, v.hourly_rate, s.date, s.start_time, s.end_time
		FROM vehicles v
		INNER JOIN schedules s ON v.vehicle_id = s.vehicle_id
		WHERE s.schedule_id = ?;
	`
	err := db.QueryRow(query, scheduleId).Scan(&vehicle.ScheduleID, &vehicle.VehicleID, &vehicle.Type, &vehicle.Brand, &vehicle.Model, &vehicle.LicensePlate, &vehicle.HourlyRate, &vehicle.Date, &vehicle.StartTime, &vehicle.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Vehicle not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Calculate the base cost of the vehicle
	startTime, err := time.Parse("15:04:05", vehicle.StartTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse start time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	endTime, err := time.Parse("15:04:05", vehicle.EndTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse end time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	baseCost := vehicle.HourlyRate * endTime.Sub(startTime).Hours()
	vehicle.BaseCost = baseCost

	w.WriteHeader(http.StatusOK)
	response := Response{"Vehicle found", &vehicle}
	json.NewEncoder(w).Encode(response)
}

// Get all the booking that are confirmed and greater than today's date
func getRentalHistory(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Variable to hold the rental history
	var history []VehicleBookingDetails
	// Get the user email from the URL
	userId := mux.Vars(r)["id"]
	// Struct for response
	type Response struct {
		Message  string                  `json:"message"`
		Vehicles []VehicleBookingDetails `json:"vehicles"`
	}
	// Validate user before proceeding
	_, err := validateUser(userId)
	if err != nil {
		// If user validation fails, send an error response
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", []VehicleBookingDetails{}}
		json.NewEncoder(w).Encode(response)
		return
	}
	// SQL query to get rental history for the user
	query := `
		SELECT 
		b.booking_id,
		b.schedule_id,
		b.user_id,
		b.status,
		v.type,
		v.brand,
		v.model AS vehicle_model,
		v.license_plate,
		s.date AS schedule_date,
		s.start_time,
		s.end_time
	FROM bookings b
	JOIN schedules s ON b.schedule_id = s.schedule_id
	JOIN vehicles v ON s.vehicle_id = v.vehicle_id
	WHERE b.user_id = ? 
	AND b.status = 'Completed'
	ORDER BY s.date DESC, s.start_time DESC, s.end_time DESC;`

	rows, err := db.Query(query, userId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record VehicleBookingDetails
		err := rows.Scan(&record.BookingID, &record.ScheduleID, &record.UserID, &record.Status, &record.Type, &record.Brand, &record.Model, &record.LicensePlate, &record.ScheduleDate, &record.StartTime, &record.EndTime)
		if err != nil {
			http.Error(w, "Error reading vehicle data", http.StatusInternalServerError)
			log.Println("Error reading vehicle data:", err)
			return
		}
		history = append(history, record)
	}
	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		log.Println("Error iterating over rows:", err)
		return
	}
	// Check if the user has any rental history
	if len(history) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No rental history found", []VehicleBookingDetails{}}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send the rental history as a JSON response
	w.WriteHeader(http.StatusOK)
	response := Response{"Rental history found", history}
	json.NewEncoder(w).Encode(response)
}

// Get all the booking that are confirmed and greater than today's date
func getUpcomingRental(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Variable to hold the rental history
	var history []VehicleBookingDetails
	// Get the user email from the URL
	userId := mux.Vars(r)["id"]
	// Struct for response
	type Response struct {
		Message  string                  `json:"message"`
		Vehicles []VehicleBookingDetails `json:"vehicles"`
	}
	// Validate user before proceeding
	_, err := validateUser(userId)
	if err != nil {
		// If user validation fails, send an error response
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", []VehicleBookingDetails{}}
		json.NewEncoder(w).Encode(response)
		return
	}
	// SQL query to get rental history for the user
	query := `SELECT b.booking_id, b.schedule_id, b.user_id, b.status, b.base_cost, b.promo_code, b.membership_discount, b.promotion_discount, b.discount_applied, b.total_amount,
			v.type, v.brand, v.model, v.license_plate, s.date AS schedule_date, s.start_time, s.end_time, v.hourly_rate
			  FROM bookings b
			  JOIN schedules s ON b.schedule_id = s.schedule_id
			  JOIN vehicles v ON s.vehicle_id = v.vehicle_id
			  WHERE b.user_id = ? AND b.status = 'Confirmed' AND s.date >= CURDATE()
			  ORDER BY s.date ASC, s.start_time ASC, s.end_time ASC;`

	rows, err := db.Query(query, userId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record VehicleBookingDetails
		err := rows.Scan(&record.BookingID, &record.ScheduleID, &record.UserID, &record.Status, &record.BaseCost, &record.PromotionCode, &record.MembershipDiscount, &record.PromotionDiscount, &record.DiscountApplied, &record.TotalAmount, &record.Type, &record.Brand, &record.Model, &record.LicensePlate, &record.ScheduleDate, &record.StartTime, &record.EndTime, &record.HourlyRate)
		if err != nil {
			http.Error(w, "Error reading vehicle data", http.StatusInternalServerError)
			log.Println("Error reading vehicle data:", err)
			return
		}
		history = append(history, record)
	}
	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		log.Println("Error iterating over rows:", err)
		return
	}
	// Check if the user has any rental history
	if len(history) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No upcoming rentals found", []VehicleBookingDetails{}}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send the rental history as a JSON response
	w.WriteHeader(http.StatusOK)
	response := Response{"Upcoming rentals found", history}
	json.NewEncoder(w).Encode(response)
}

// Create a booking session for the user
func createBookingSession(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string                 `json:"message"`
		Booking *VehicleBookingDetails `json:"booking"`
	}

	// Get user_id and schedule_id from the URL parameters
	userID := mux.Vars(r)["id"]
	scheduleID := mux.Vars(r)["scheduleId"]

	// Validate user ID
	user, err := validateUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate license expiry
	licenseExpiry, err := time.Parse("2006-01-02", user.LicenseExpiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse license expiry date", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fetch schedule date and reservation status in a single query
	var scheduleDate string
	var isReserved bool
	var vehicleHourlyRate float64
	var startTime, endTime string
	query := `SELECT s.date, s.is_reserved, s.start_time, s.end_time, v.hourly_rate
			  FROM schedules s INNER JOIN vehicles v ON s.vehicle_id = v.vehicle_id WHERE schedule_id = ?`
	err = db.QueryRow(query, scheduleID).Scan(&scheduleDate, &isReserved, &startTime, &endTime, &vehicleHourlyRate)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Schedule not found", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to query schedule", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Compare the license expiry date with the schedule date
	scheduleDateTime, err := time.Parse("2006-01-02", scheduleDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse schedule date", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	if licenseExpiry.Before(scheduleDateTime) {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"User's license is expired", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if the schedule is already reserved
	if isReserved {
		w.WriteHeader(http.StatusConflict)
		response := Response{"Schedule is already reserved", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if user exceeded their booking limit
	membership, err := getMembershipDetails(user.MembershipId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"Membership not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	var bookingCount int
	countQuery := `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE user_id = ? AND (status = 'Confirmed' OR status = 'Completed')
	`
	err = db.QueryRow(countQuery, userID).Scan(&bookingCount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to query booking count", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	if bookingCount >= membership.BookingLimit {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Message: fmt.Sprintf("You have exceeded the booking limit of %d", membership.BookingLimit),
			Booking: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate the amount using the calculateAmount function
	membershipDiscount := membership.HourlyRateDiscount
	startTimeFmt, err := time.Parse("15:04:05", startTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse start time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	endTimeFmt, err := time.Parse("15:04:05", endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse end time", nil}
		json.NewEncoder(w).Encode(response)
	}
	baseAmount, membershipDiscount, promotionDiscount, totalDiscount, totalAmount, err := calculateAmount(vehicleHourlyRate, startTimeFmt, endTimeFmt, membershipDiscount, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to calculate amount", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create the booking
	insertQuery := `
		INSERT INTO bookings (schedule_id, user_id, status, base_cost, membership_discount, promotion_discount, discount_applied, total_amount)
		VALUES (?, ?, 'Pending', ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(insertQuery, scheduleID, userID, baseAmount, membershipDiscount, promotionDiscount, totalDiscount, totalAmount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to create booking", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the new booking ID
	bookingId, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to retrieve booking ID", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Mark schedule as reserved
	updateQuery := `UPDATE schedules SET is_reserved = TRUE WHERE schedule_id = ?`
	_, err = db.Exec(updateQuery, scheduleID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update schedule", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fetch booking details
	var bookingDetails VehicleBookingDetails
	selectQuery := `
		SELECT 
			b.booking_id, b.schedule_id, b.user_id, b.status, b.base_cost, b.membership_discount, b.promotion_discount, b.discount_applied, b.total_amount,
			v.type, v.brand, v.model, v.license_plate, 
			s.date AS schedule_date, s.start_time, s.end_time
		FROM bookings b
		INNER JOIN schedules s ON b.schedule_id = s.schedule_id
		INNER JOIN vehicles v ON s.vehicle_id = v.vehicle_id
		WHERE b.booking_id = ?
	`
	err = db.QueryRow(selectQuery, bookingId).Scan(
		&bookingDetails.BookingID,
		&bookingDetails.ScheduleID,
		&bookingDetails.UserID,
		&bookingDetails.Status,
		&bookingDetails.BaseCost,
		&bookingDetails.MembershipDiscount,
		&bookingDetails.PromotionDiscount,
		&bookingDetails.DiscountApplied,
		&bookingDetails.TotalAmount,
		&bookingDetails.Type,
		&bookingDetails.Brand,
		&bookingDetails.Model,
		&bookingDetails.LicensePlate,
		&bookingDetails.ScheduleDate,
		&bookingDetails.StartTime,
		&bookingDetails.EndTime,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to fetch booking details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Respond with booking details
	w.WriteHeader(http.StatusCreated)
	response := Response{Message: "Booking session created successfully", Booking: &bookingDetails}
	json.NewEncoder(w).Encode(response)
}

// update total amount of the booking if promotion code provided
func addPromotionCode(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message string                 `json:"message"`
		Booking *VehicleBookingDetails `json:"booking"`
	}

	// Extract user ID, booking ID, and promo code from the URL
	userID := mux.Vars(r)["id"]
	bookingIDStr := mux.Vars(r)["bookingId"]
	promoCode := mux.Vars(r)["promoCode"]

	// Validate user
	user, err := validateUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse booking ID
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Invalid booking ID format", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get booking details
	var vehicleHourlyRate float64
	var startTime, endTime string
	query := `
        SELECT v.hourly_rate, s.start_time, s.end_time
        FROM bookings b
        INNER JOIN schedules s ON b.schedule_id = s.schedule_id
        INNER JOIN vehicles v ON s.vehicle_id = v.vehicle_id
        WHERE b.booking_id = ? AND b.user_id = ? AND b.status = 'Pending'
    `
	err = db.QueryRow(query, bookingID, userID).Scan(&vehicleHourlyRate, &startTime, &endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Booking session not found or not in 'Pending' status", nil}
			json.NewEncoder(w).Encode(response)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Failed to fetch booking details", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	// Parse start and end times
	startTimeFmt, err := time.Parse("15:04:05", startTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse start time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	endTimeFmt, err := time.Parse("15:04:05", endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse end time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	membership, err := getMembershipDetails(user.MembershipId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"Membership not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	membershipDiscount := membership.HourlyRateDiscount

	// Calculate the new total amount after applying the promotion discount
	_, _, promotionDiscountAmt, totalDiscountAmt, totalAmt, err := calculateAmount(vehicleHourlyRate, startTimeFmt, endTimeFmt, membershipDiscount, promoCode)
	if err != nil {
		if err.Error() == "promo code not found" {
			w.WriteHeader(http.StatusBadRequest) // Use 400 for client-side error
			response := Response{"Promo Code Not Found", nil}
			json.NewEncoder(w).Encode(response)
			return
		} else if err.Error() == "promo code not valid" {
			w.WriteHeader(http.StatusBadRequest) // Use 400 for client-side error
			response := Response{"Promo Code Not Valid", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		// For other errors, use internal server error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to calculate amount", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the booking with the new total amount
	updateQuery := `
		UPDATE bookings 
		SET promo_code = ?, promotion_discount = ?, discount_applied = ?, total_amount = ? 
		WHERE booking_id = ? AND user_id = ? AND status = 'Pending'
	`
	_, err = db.Exec(updateQuery, promoCode, promotionDiscountAmt, totalDiscountAmt, totalAmt, bookingID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update booking", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fetch booking details
	var bookingDetails VehicleBookingDetails
	selectQuery := `
		SELECT 
			b.booking_id, b.schedule_id, b.user_id, b.status, b.base_cost, b.promo_code, b.membership_discount, b.promotion_discount, b.discount_applied, b.total_amount,
			v.type, v.brand, v.model, v.license_plate, 
			s.date AS schedule_date, s.start_time, s.end_time
		FROM bookings b
		INNER JOIN schedules s ON b.schedule_id = s.schedule_id
		INNER JOIN vehicles v ON s.vehicle_id = v.vehicle_id
		WHERE b.booking_id = ?
	`
	err = db.QueryRow(selectQuery, bookingID).Scan(
		&bookingDetails.BookingID,
		&bookingDetails.ScheduleID,
		&bookingDetails.UserID,
		&bookingDetails.Status,
		&bookingDetails.BaseCost,
		&bookingDetails.PromotionCode,
		&bookingDetails.MembershipDiscount,
		&bookingDetails.PromotionDiscount,
		&bookingDetails.DiscountApplied,
		&bookingDetails.TotalAmount,
		&bookingDetails.Type,
		&bookingDetails.Brand,
		&bookingDetails.Model,
		&bookingDetails.LicensePlate,
		&bookingDetails.ScheduleDate,
		&bookingDetails.StartTime,
		&bookingDetails.EndTime,
	)
	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to fetch booking details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{"Promotion code applied successfully", &bookingDetails}
	json.NewEncoder(w).Encode(response)
}

func deleteBookingSession(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message   string `json:"message"`
		BookingID *int64 `json:"booking_id"` // Nullable BookingID
	}

	// Get user_id and booking_id from the URL parameters
	userID := mux.Vars(r)["id"]
	bookingIDStr := mux.Vars(r)["bookingId"]

	// Validate user before proceeding
	_, err := validateUser(userID)
	if err != nil {
		// If user validation fails, send an error response
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert bookingID from string to int64 using strconv.ParseInt
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64) // base 10, 64-bit
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Invalid booking ID format", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get the booking session for the user
	var scheduleID int64
	selectQuery := `
		SELECT schedule_id
		FROM bookings
		WHERE booking_id = ? AND user_id = ? AND status = 'Pending'
	`
	err = db.QueryRow(selectQuery, bookingID, userID).Scan(&scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Booking session not found or not in 'Pending' status", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to query booking", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the booking status to "SessionExpired"
	updateBookingQuery := `
		UPDATE bookings
		SET status = 'SessionExpired'
		WHERE booking_id = ?
	`
	_, err = db.Exec(updateBookingQuery, bookingID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update booking status", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the schedule to mark it as not reserved
	updateScheduleQuery := `UPDATE schedules SET is_reserved = FALSE WHERE schedule_id = ?`
	_, err = db.Exec(updateScheduleQuery, scheduleID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update schedule", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	response := Response{
		Message:   "Booking session expired and schedule updated successfully",
		BookingID: &bookingID, // Return the booking ID
	}
	json.NewEncoder(w).Encode(response)
}

// Delete the booking that are confirmed
func deleteBooking(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Get the booking_id and user_id from the URL
	bookingId := mux.Vars(r)["bookingId"]
	userId := mux.Vars(r)["id"]

	// Struct for response
	type Response struct {
		Message string `json:"message"`
	}
	// Validate user before proceeding
	_, err := validateUser(userId)
	if err != nil {
		// If user validation fails, send an error response
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get the booking details to check status and timing
	query := `
        SELECT b.status, s.date, s.start_time
        FROM bookings b
        JOIN schedules s ON b.schedule_id = s.schedule_id
        WHERE b.booking_id = ? AND b.user_id = ?;
    `

	var status, scheduledDate string
	var scheduledTime string

	// Execute the query to retrieve booking details
	err = db.QueryRow(query, bookingId, userId).Scan(&status, &scheduledDate, &scheduledTime)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{Message: "Booking not found"}
			json.NewEncoder(w).Encode(response)
		} else {
			fmt.Println(err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check if the status is 'Confirmed'
	if status != "Confirmed" {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{Message: "Only confirmed bookings can be cancelled"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse the scheduled date and time as SGT
	timezone := "Asia/Singapore"
	loc, err := time.LoadLocation(timezone) // Load Singapore Timezone
	if err != nil {
		// Handle error
		http.Error(w, "Error loading Singapore timezone", http.StatusInternalServerError)
		return
	}

	// Combine the scheduled date and time into one formatted string
	scheduledTimeFormatted := fmt.Sprintf("%s %s", scheduledDate, scheduledTime)
	scheduledDatetime, err := time.ParseInLocation("2006-01-02 15:04:05", scheduledTimeFormatted, loc)
	if err != nil {
		// Handle error
		http.Error(w, "Error parsing scheduled time", http.StatusInternalServerError)
		return
	}

	// Get the current time in Singapore Time (SGT)
	currentTime := time.Now().In(loc)

	// Calculate the time remaining
	timeRemaining := scheduledDatetime.Sub(currentTime)

	fmt.Println("Time Remaining:", timeRemaining)
	fmt.Println("Scheduled Time in SGT:", scheduledDatetime)
	fmt.Println("Current Time in SGT:", currentTime)

	// Check if the cancellation is within 24 hours
	if timeRemaining < 24*time.Hour {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{Message: "Cancellation must be done at least 24 hours before the booking"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// SQL query to update booking status and related schedule
	cancelQueryStep1 := `
		UPDATE bookings
		SET status = 'Cancelled'
		WHERE booking_id = ? AND user_id = ?;
	`

	cancelQueryStep2 := `
		UPDATE schedules
		SET is_reserved = FALSE
		WHERE schedule_id = (
			SELECT schedule_id
			FROM bookings
			WHERE booking_id = ? AND user_id = ?
		);
	`
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Step 1: Update the booking status
	_, err = tx.Exec(cancelQueryStep1, bookingId, userId)
	if err != nil {
		tx.Rollback() // Roll back the transaction on error
		log.Fatalf("Failed to update booking status: %v", err)
	}

	// Step 2: Update the schedule's is_reserved field
	_, err = tx.Exec(cancelQueryStep2, bookingId, userId)
	if err != nil {
		tx.Rollback() // Roll back the transaction on error
		log.Fatalf("Failed to update schedule: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	// Send the response as a JSON message
	w.WriteHeader(http.StatusOK)
	response := Response{Message: "Booking cancelled successfully"}
	json.NewEncoder(w).Encode(response)
}

// Verify the booking exists and is in 'Pending' status
func verifyBooking(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Get the booking_id and user_id from the URL
	userId := mux.Vars(r)["id"]
	bookingId := mux.Vars(r)["bookingId"]

	// Struct for response
	type Response struct {
		Message string                 `json:"message"`
		Booking *VehicleBookingDetails `json:"booking"`
	}

	// Validate user before proceeding
	_, err := validateUser(userId)
	if err != nil {
		// If user validation fails, send an error response
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get the booking details to check status
	query := `
		SELECT b.booking_id, b.schedule_id, b.user_id, b.status, b.base_cost, b.promo_code, b.membership_discount, b.promotion_discount, b.discount_applied, b.total_amount,
		v.type, v.brand, v.model, v.license_plate, s.date AS schedule_date, s.start_time, s.end_time
		FROM bookings b
		JOIN schedules s ON b.schedule_id = s.schedule_id
		JOIN vehicles v ON s.vehicle_id = v.vehicle_id
		WHERE b.booking_id = ? AND b.user_id = ? AND b.status = 'Pending';
	`
	// Execute the query to retrieve booking details
	var bookingDetails VehicleBookingDetails
	err = db.QueryRow(query, bookingId, userId).Scan(
		&bookingDetails.BookingID,
		&bookingDetails.ScheduleID,
		&bookingDetails.UserID,
		&bookingDetails.Status,
		&bookingDetails.BaseCost,
		&bookingDetails.PromotionCode,
		&bookingDetails.MembershipDiscount,
		&bookingDetails.PromotionDiscount,
		&bookingDetails.DiscountApplied,
		&bookingDetails.TotalAmount,
		&bookingDetails.Type,
		&bookingDetails.Brand,
		&bookingDetails.Model,
		&bookingDetails.LicensePlate,
		&bookingDetails.ScheduleDate,
		&bookingDetails.StartTime,
		&bookingDetails.EndTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Booking not found or not in 'Pending' status", nil}
			json.NewEncoder(w).Encode(response)
		} else {
			fmt.Println(err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{Message: "Booking found", Booking: &bookingDetails}
	json.NewEncoder(w).Encode(response)
}

// Handler for the /booking confirmation endpoint to handle the payment confirmation, chnage the booking status to 'Confirmed'
func confirmBooking(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Get the booking_id and user_id from the URL
	userId := mux.Vars(r)["id"]
	bookingId := mux.Vars(r)["bookingId"]

	// Struct for response
	type Response struct {
		Message string `json:"message"`
	}

	// Parse the incoming JSON payload to check if payment was successful
	var paymentInfo struct {
		PaymentSuccess bool `json:"paymentSuccess"`
	}
	err := json.NewDecoder(r.Body).Decode(&paymentInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Invalid payment confirmation data"}
		json.NewEncoder(w).Encode(response)
		return
	}
	fmt.Println(paymentInfo)

	// If payment was not successful, don't confirm the booking
	if !paymentInfo.PaymentSuccess {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Payment failed, cannot confirm booking"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate user before proceeding
	_, err = validateUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get the booking details to check status
	query := `
        SELECT status
        FROM bookings
        WHERE booking_id = ? AND user_id = ?
    `
	// Execute the query to retrieve booking details
	var status string
	err = db.QueryRow(query, bookingId, userId).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Booking not found or not in 'Pending' status"}
			json.NewEncoder(w).Encode(response)
		} else {
			fmt.Println(err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	if status != "Pending" {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Only pending bookings can be confirmed"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the booking status to 'Confirmed'
	updateQuery := `
        UPDATE bookings
        SET status = 'Confirmed'
        WHERE booking_id = ? AND user_id = ?;
    `
	_, err = db.Exec(updateQuery, bookingId, userId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send the response as a JSON message
	w.WriteHeader(http.StatusOK)
	response := Response{Message: "Booking confirmed successfully"}
	json.NewEncoder(w).Encode(response)
}

// Get Get selected vehicle details by hourly_rate
func getVehicleDetailsByHourlyRate(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Variable to hold the vehicle details
	var vehicles []VehicleSchedules
	// Get the hourly_rate from the URL
	hourlyRate := mux.Vars(r)["hourlyRate"]
	// Struct for response
	type Response struct {
		Message string             `json:"message"`
		Vehicle []VehicleSchedules `json:"vehicles"`
	}
	// Query to get the vehicle details
	query := `
		SELECT s.schedule_id, v.vehicle_id, v.type, v.brand, v.model, v.license_plate, v.hourly_rate, s.date, s.start_time, s.end_time
		FROM vehicles v
		INNER JOIN schedules s ON v.vehicle_id = s.vehicle_id
		WHERE v.hourly_rate = ? AND s.date >= CURDATE() AND s.is_reserved = FALSE;
	`
	rows, err := db.Query(query, hourlyRate)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var vehicle VehicleSchedules
		err := rows.Scan(&vehicle.ScheduleID, &vehicle.VehicleID, &vehicle.Type, &vehicle.Brand, &vehicle.Model, &vehicle.LicensePlate, &vehicle.HourlyRate, &vehicle.Date, &vehicle.StartTime, &vehicle.EndTime)
		if err != nil {
			http.Error(w, "Error reading vehicle data", http.StatusInternalServerError)
			log.Println("Error reading vehicle data:", err)
			return
		}
		vehicle.BaseCost = vehicle.HourlyRate
		vehicles = append(vehicles, vehicle)
	}
	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		log.Println("Error iterating over rows:", err)
		return
	}
	// If no vehicles were found, return a "Not Found" message
	if len(vehicles) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No vehicle available in the given range", vehicles}
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{"Vehicles available in the given range", vehicles}
	json.NewEncoder(w).Encode(response)
}

// Update the  booking details by booking id
func updateBooking(w http.ResponseWriter, r *http.Request) {
	// Set the header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Get the booking_id and user_id from the URL
	userId := mux.Vars(r)["id"]
	bookingId := mux.Vars(r)["bookingId"]
	scheduleId := mux.Vars(r)["scheduleId"]
	// Struct for response
	type Response struct {
		Message               string                 `json:"message"`
		VehicleBookingDetails *VehicleBookingDetails `json:"booking"`
	}
	// Validate user before proceeding
	_, err := validateUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := Response{"User not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Query to get the booking details to check status
	query := `SELECT s.schedule_id, s.start_time, s.end_time, s.date, b.status, v.hourly_rate
	FROM bookings b
	JOIN schedules s ON b.schedule_id = s.schedule_id
	JOIN vehicles v ON s.vehicle_id = v.vehicle_id
	WHERE booking_id = ? AND user_id = ?`
	// Execute the query to retrieve booking details
	var bookedstatus string
	var bookedscheduleId int64
	var bookedvehicleHourlyRate float64
	var bookedscheduleStartTime, bookedscheduleEndTime string
	var bookedscheduleDate string
	err = db.QueryRow(query, bookingId, userId).Scan(&bookedscheduleId, &bookedscheduleStartTime, &bookedscheduleEndTime, &bookedscheduleDate, &bookedstatus, &bookedvehicleHourlyRate)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Booking not found", nil}
			json.NewEncoder(w).Encode(response)
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Database error", nil}
			json.NewEncoder(w).Encode(response)
		}
		return
	}
	if bookedstatus != "Confirmed" {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Only confirmed bookings can be updated", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Cannot update withtin 24 hours of the booking
	timezone := "Asia/Singapore"
	loc, err := time.LoadLocation(timezone) // Load Singapore Timezone
	if err != nil {
		// Handle error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error loading Singapore timezone", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Combine the booked schedule date and start time into one formatted string
	scheduledTimeFormatted := fmt.Sprintf("%s %s", bookedscheduleDate, bookedscheduleStartTime)
	bookedscheduleDateTime, err := time.ParseInLocation("2006-01-02 15:04:05", scheduledTimeFormatted, loc)
	if err != nil {
		// Handle error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse schedule date", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the current time in Singapore Time (SGT)
	currentTime := time.Now().In(loc)

	// Check if the cancellation is within 24 hours from now
	if currentTime.Add(24 * time.Hour).After(bookedscheduleDateTime) {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Cannot update booking within 24 hours of the booking", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query to get the schedule details
	query = `SELECT s.date, s.start_time, s.end_time, v.hourly_rate, s.is_reserved
			 FROM schedules s JOIN vehicles v ON s.vehicle_id = v.vehicle_id
			 WHERE schedule_id = ? `
	// Execute the query to retrieve schedule details
	var date string
	var scheduleStartTime, scheduleEndTime string
	var vehicleHourlyRate float64
	var scheduleStatus bool
	err = db.QueryRow(query, scheduleId).Scan(&date, &scheduleStartTime, &scheduleEndTime, &vehicleHourlyRate, &scheduleStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Schedule not found or already reserved", nil}
			json.NewEncoder(w).Encode(response)
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Database error", nil}
			json.NewEncoder(w).Encode(response)
		}
		return
	}
	// Check if the schedule is already reserved
	if scheduleStatus {
		w.WriteHeader(http.StatusConflict)
		response := Response{"Schedule is already reserved", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check if the booking is for the same vehicle type
	if bookedvehicleHourlyRate != vehicleHourlyRate {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Booking is for a different vehicle type", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Check duration of the booking
	bookedscheduleStartTimeFmt, err := time.Parse("15:04:05", bookedscheduleStartTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse start time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	bookedscheduleEndTimeFmt, err := time.Parse("15:04:05", bookedscheduleEndTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse end time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	scheduleStartTimeFmt, err := time.Parse("15:04:05", scheduleStartTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse start time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	scheduleEndTimeFmt, err := time.Parse("15:04:05", scheduleEndTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to parse end time", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Calculate the duration
	bookedDuration := bookedscheduleEndTimeFmt.Sub(bookedscheduleStartTimeFmt)
	newDuration := scheduleEndTimeFmt.Sub(scheduleStartTimeFmt)
	if bookedDuration != newDuration {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{"Booking duration does not match the schedule duration", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Update the booking with the new schedule details
	updateQuery := `UPDATE bookings SET schedule_id = ? WHERE booking_id = ? AND user_id = ?`
	_, err = db.Exec(updateQuery, scheduleId, bookingId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update booking", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Update the schedule to mark it as reserved
	updateScheduleQuery := `UPDATE schedules SET is_reserved = TRUE WHERE schedule_id = ?`
	_, err = db.Exec(updateScheduleQuery, scheduleId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update schedule", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Update the schedule to mark it as not reserved
	updateScheduleQuery = `UPDATE schedules SET is_reserved = FALSE WHERE schedule_id = ?`
	_, err = db.Exec(updateScheduleQuery, bookedscheduleId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to update schedule", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Fetch booking details
	var bookingDetails VehicleBookingDetails
	selectQuery := `SELECT b.booking_id, b.schedule_id, b.user_id, b.status, b.base_cost, b.promo_code, b.membership_discount, b.promotion_discount, b.discount_applied, b.total_amount, v.type, v.brand, v.model, v.license_plate, s.date AS schedule_date, s.start_time, s.end_time 
					FROM bookings b JOIN schedules s ON b.schedule_id = s.schedule_id 
					JOIN vehicles v ON s.vehicle_id = v.vehicle_id WHERE b.booking_id = ?`
	err = db.QueryRow(selectQuery, bookingId).Scan(&bookingDetails.BookingID, &bookingDetails.ScheduleID, &bookingDetails.UserID, &bookingDetails.Status, &bookingDetails.BaseCost, &bookingDetails.PromotionCode, &bookingDetails.MembershipDiscount, &bookingDetails.PromotionDiscount, &bookingDetails.DiscountApplied, &bookingDetails.TotalAmount, &bookingDetails.Type, &bookingDetails.Brand, &bookingDetails.Model, &bookingDetails.LicensePlate, &bookingDetails.ScheduleDate, &bookingDetails.StartTime, &bookingDetails.EndTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Failed to fetch booking details", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Response{"Booking updated successfully", &bookingDetails}
	json.NewEncoder(w).Encode(response)
}
