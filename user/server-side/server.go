package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"math/rand"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

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

var db *sql.DB

// Initialise the user_svc_db database connection
func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/user_svc_db")
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
	router.HandleFunc("/api/v1/register", registerUser).Methods("POST")
	router.HandleFunc("/api/v1/verify", verifyUser).Methods("POST")
	router.HandleFunc("/api/v1/login", loginUser).Methods("POST")
	router.HandleFunc("/api/v1/user/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/api/v1/user/{id}", getUser).Methods("GET")
	router.HandleFunc("/api/v1/password", updatePassword).Methods("PUT")
	router.HandleFunc("/api/v1/validate-user/{id}", userExists).Methods("GET")
	router.HandleFunc("/api/v1/membership/{id}", getMembership).Methods("GET")
	handler := cors.Default().Handler(router)
	fmt.Println("Listening at port 8000")
	log.Fatal(http.ListenAndServe(":8000", handler))
}

// Hash the password using bcrypt
func hashPassword(password string) (string, error) {
	// Generate a hashed password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Calculate age from dob
func calculateAge(dobStr string) (int, error) {
	// Parse the dob string into time.Time
	dob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format")
	}

	// Calculate the age
	currentTime := time.Now()
	age := currentTime.Year() - dob.Year()
	if currentTime.YearDay() < dob.YearDay() {
		age--
	}

	return age, nil
}

// Creating a post function to register User
func registerUser(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")

	// Response struct for user registration
	type RegisterResponse struct {
		Message          string `json:"message"`
		VerificationCode string `json:"verification_code"`
		User             User   `json:"user"`
	}

	// Create a new instance of User struct
	var newUser User

	// Read the request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := RegisterResponse{
			Message: "Failed to read request body",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into the newUser struct
	err = json.Unmarshal(jsonByte, &newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := RegisterResponse{
			Message: "Invalid user data",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if the user's age is greater than 18 by using the function calculateAge
	age, err := calculateAge(newUser.Dob)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := RegisterResponse{
			Message: "Invalid date format",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if age < 18 {
		w.WriteHeader(http.StatusForbidden)
		response := RegisterResponse{
			Message: "User must be 18 or older",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate license expiry date
	licenseExpiry, err := time.Parse("2006-01-02", newUser.LicenseExpiry)
	if err != nil || licenseExpiry.Before(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		response := RegisterResponse{
			Message: "Invalid or expired license date",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Hash the password by calling the function hashPassword
	hashedPassword, err := hashPassword(newUser.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := RegisterResponse{
			Message: "Failed to hash password",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Assign user membership to be default value (Basic)
	newUser.MembershipId = "Basic"

	// Assign a random verification code to the user
	verificationCode := strconv.Itoa(rand.Intn(1000000))

	// Insert the user data into the user_svc_db
	query := `INSERT INTO users (name, email, phone, dob, password, membership_id, license_number, license_expiry, verification_code, verified) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, newUser.Name, newUser.Email, newUser.Phone, newUser.Dob, hashedPassword, newUser.MembershipId, newUser.LicenseNumber, newUser.LicenseExpiry, verificationCode, false)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			w.WriteHeader(http.StatusConflict)
			response := RegisterResponse{
				Message: "Email or phone number already exists",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := RegisterResponse{
			Message: "Failed to insert user into database",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Retrieve the last inserted user_id
	userID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := RegisterResponse{
			Message: "Failed to retrieve user ID",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	fmt.Println(userID)
	// Set the user ID in the newUser struct
	newUser.UserID = int(userID)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	response := RegisterResponse{
		Message:          "User registered successfully",
		VerificationCode: verificationCode,
		User:             newUser,
	}
	json.NewEncoder(w).Encode(response)
}

// Creating a post function to verify authentication
func verifyUser(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")
	var verificationRequest struct {
		Email            string `json:"email"`
		VerificationCode string `json:"verification_code"`
	}
	type LoginResponse struct {
		Message string `json:"message"`
		UserId  int    `json:"user_id"`
	}
	// Read the request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into the verificationRequest struct
	err = json.Unmarshal(jsonByte, &verificationRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := LoginResponse{
			Message: "Invalid verification data",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Retrieve the user by email
	var user User
	var userId int
	query := `SELECT user_id, email, verification_code, verified FROM users WHERE email = ?`
	err = db.QueryRow(query, verificationRequest.Email).Scan(&userId, &user.Email, &user.VerificationCode, &user.Verified)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := LoginResponse{
				Message: "User not found",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	// Check if the user is already verified
	if user.Verified {
		w.WriteHeader(http.StatusConflict)
		response := LoginResponse{
			Message: "User with the email of " + user.Email + " is already verified",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if the verification code matches
	if user.VerificationCode != verificationRequest.VerificationCode {
		// Respond with unauthorized if the verification code does not match
		w.WriteHeader(http.StatusUnauthorized)
		response := LoginResponse{
			Message: "Invalid verification code",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		// Update the user verification status if the verification code matches
		_, err = db.Exec("UPDATE users SET verified = ? WHERE email = ?", true, verificationRequest.Email)
		if err != nil {
			http.Error(w, "Failed to update user verification status", http.StatusInternalServerError)
			return
		}
		// Respond with success
		w.WriteHeader(http.StatusOK)
		respsonse := struct {
			Message string `json:"message"`
			UserId  int    `json:"user_id"`
		}{
			Message: "User verified successfully",
			UserId:  userId,
		}
		json.NewEncoder(w).Encode(respsonse)
	}
}

// Creating a post function to login user
func loginUser(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")
	// LoginRequest struct for login data
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type LoginResponse struct {
		Message string `json:"message"`
		UserId  int    `json:"user_id"`
	}

	// Read the request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into the verificationRequest struct
	err = json.Unmarshal(jsonByte, &loginRequest)
	if err != nil {
		http.Error(w, "Invalid login data", http.StatusBadRequest)
		return
	}
	var hashedPassword string
	var userId int
	// Retrieve the user by email
	var user User
	query := `SELECT user_id, email, password, verified FROM users WHERE email = ?`
	err = db.QueryRow(query, loginRequest.Email).Scan(&userId, &user.Email, &hashedPassword, &user.Verified)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := LoginResponse{
				Message: "User not found",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	// Check if the user is verified
	if !user.Verified {
		w.WriteHeader(http.StatusForbidden)
		response := LoginResponse{
			Message: "User with the email of " + user.Email + " is not verified",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Compare the password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginRequest.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := LoginResponse{
			Message: "Invalid password",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Successful login
	w.WriteHeader(http.StatusOK)
	response := LoginResponse{
		Message: "User logged in successfully",
		UserId:  userId,
	}
	json.NewEncoder(w).Encode(response)
}

// Create a function to update user details
func updateUser(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")

	// Response struct for user registration
	type UpdateResponse struct {
		Message          string `json:"message"`
		VerificationCode string `json:"verification_code"`
		User             User   `json:"user"`
	}

	// Get user ID from URL params (assuming it's passed)
	userId := mux.Vars(r)["id"]

	// Validate the user email is found in db
	var currentemail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := db.QueryRow(query, userId).Scan(&currentemail)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := UpdateResponse{
				Message: "User not found",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	// Create a new instance of User struct for updated details
	var updatedUser User

	// Read the request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := UpdateResponse{
			Message: "Failed to read request body",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into the updateRequest struct
	err = json.Unmarshal(jsonByte, &updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := UpdateResponse{
			Message: "Invalid update data",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate new verification code if email is updated
	var verificationCode string
	setVerificationCode := false

	// Check if the email is being updated
	if updatedUser.Email != "" && updatedUser.Email != currentemail {
		// Generate a new verification code if the email has changed
		verificationCode = strconv.Itoa(rand.Intn(1000000))
		setVerificationCode = true
	}

	// Validate license expiry date if provided
	if updatedUser.LicenseExpiry != "" {
		licenseExpiry, err := time.Parse("2006-01-02", updatedUser.LicenseExpiry)
		if err != nil || licenseExpiry.Before(time.Now()) {
			w.WriteHeader(http.StatusBadRequest)
			response := UpdateResponse{
				Message: "Invalid or expired license date",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// SQL query for updating user details
	updateQuery := `
	UPDATE users 
	SET 
		email = COALESCE(NULLIF(?, ''), email),
		name = COALESCE(NULLIF(?, ''), name),
		phone = COALESCE(NULLIF(?, ''), phone),
		membership_id = COALESCE(NULLIF(?, ''), membership_id),
		license_number = COALESCE(NULLIF(?, ''), license_number),
		license_expiry = COALESCE(NULLIF(?, ''), license_expiry)`
	if setVerificationCode {
		updateQuery += `, verification_code = ?, verified = ?`
	}
	updateQuery += ` WHERE user_id = ?`

	// Execute the update query
	if setVerificationCode {
		_, err = db.Exec(updateQuery, updatedUser.Email, updatedUser.Name, updatedUser.Phone, updatedUser.MembershipId, updatedUser.LicenseNumber, updatedUser.LicenseExpiry, verificationCode, false, userId)
	} else {
		_, err = db.Exec(updateQuery, updatedUser.Email, updatedUser.Name, updatedUser.Phone, updatedUser.MembershipId, updatedUser.LicenseNumber, updatedUser.LicenseExpiry, userId)
	}

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		response := UpdateResponse{
			Message: "Failed to update user",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Determine which email to use for fetching the updated record
	var newEmail string
	if setVerificationCode {
		newEmail = updatedUser.Email
	} else {
		newEmail = currentemail
	}

	//Retrieve the updated user details
	var dbuser User
	query = `SELECT name, email, phone, dob, membership_id, license_number, license_expiry, verified FROM users WHERE email = ?`
	err = db.QueryRow(query, newEmail).Scan(&dbuser.Name, &dbuser.Email, &dbuser.Phone, &dbuser.Dob, &dbuser.MembershipId, &dbuser.LicenseNumber, &dbuser.LicenseExpiry, &dbuser.Verified)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		response := UpdateResponse{
			Message: "Failed to fetch updated user",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	response := UpdateResponse{
		Message: "User updated successfully",
		User:    dbuser,
	}
	if setVerificationCode {
		response.VerificationCode = verificationCode
	}
	json.NewEncoder(w).Encode(response)
}

// Create a function to update the password by email
func updatePassword(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")

	// Response struct for user registration
	type UpdatePasswordResponse struct {
		Message string `json:"message"`
	}

	// Read the request body
	var updatedUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Read and parse the request body into the struct
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := UpdatePasswordResponse{
			Message: "Failed to read request body",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into the updatedUser struct
	err = json.Unmarshal(jsonByte, &updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := UpdatePasswordResponse{
			Message: "Invalid update data",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate the user email is found in the database
	var currentHashedPassword string
	query := `SELECT password FROM users WHERE email = ?`
	err = db.QueryRow(query, updatedUser.Email).Scan(&currentHashedPassword)
	if err != nil {
		fmt.Println(err)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := UpdatePasswordResponse{
				Message: "User not found",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Check if the new password is different from the current one
	err = bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(updatedUser.Password))
	if err == nil {
		// If the passwords are the same
		w.WriteHeader(http.StatusBadRequest)
		response := UpdatePasswordResponse{
			Message: "New password cannot be the same as the current password",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Hash the new password
	newHashedPassword, err := hashPassword(updatedUser.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := UpdatePasswordResponse{
			Message: "Failed to hash password",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// SQL query for updating user password
	updateQuery := `UPDATE users SET password = ? WHERE email = ?`

	// Execute the update query
	_, err = db.Exec(updateQuery, newHashedPassword, updatedUser.Email)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		response := UpdatePasswordResponse{
			Message: "Failed to update password",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	response := UpdatePasswordResponse{
		Message: "Password updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// Create a function to get user details by ID
func getUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL params
	userId := mux.Vars(r)["id"]

	// Retrieve the user by ID
	var user User
	query := `SELECT user_id, name, email, phone, dob, membership_id, license_number, license_expiry, verified FROM users WHERE user_id = ?`
	err := db.QueryRow(query, userId).Scan(&user.UserID, &user.Name, &user.Email, &user.Phone, &user.Dob, &user.MembershipId, &user.LicenseNumber, &user.LicenseExpiry, &user.Verified)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Respond with the user details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Function to validate the user exists in the database
func userExists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Get user_id from request URL (e.g., /users/{user_id})
	userID := mux.Vars(r)["id"]
	type Response struct {
		Message string `json:"message"`
		User    User   `json:"user"`
	}
	// Query to check if the user exists
	var foundUser User
	query := `SELECT user_id, name, email, phone, dob, membership_id, license_number, license_expiry, verified FROM users WHERE user_id = ?`
	err := db.QueryRow(query, userID).Scan(&foundUser.UserID, &foundUser.Name, &foundUser.Email, &foundUser.Phone, &foundUser.Dob, &foundUser.MembershipId, &foundUser.LicenseNumber, &foundUser.LicenseExpiry, &foundUser.Verified)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{
				Message: "User not found",
			}
			json.NewEncoder(w).Encode(response)
			return
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{
				Message: "Database error",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	// Successful login
	w.WriteHeader(http.StatusOK)
	response := Response{
		Message: "User found",
		User:    foundUser,
	}
	json.NewEncoder(w).Encode(response)
}

// Create a function to retrieve membership details by ID
func getMembership(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type once at the start
	w.Header().Set("Content-Type", "application/json")
	// Get membership ID from URL params
	membershipId := mux.Vars(r)["id"]
	type Response struct {
		Message    string      `json:"message"`
		Membership *Membership `json:"membership"`
	}
	// Retrieve the membership by ID
	var membership Membership

	query := `SELECT membership_id, hourly_rate_discount, booking_limit FROM memberships WHERE membership_id = ?`
	err := db.QueryRow(query, membershipId).Scan(&membership.MembershipId, &membership.HourlyRateDiscount, &membership.BookingLimit)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			response := Response{"Membership not found", nil}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Database error", nil}
			json.NewEncoder(w).Encode(response)
		}
		return
	}

	// Respond with the membership details
	w.WriteHeader(http.StatusOK)
	response := Response{Message: "Membership found", Membership: &membership}
	json.NewEncoder(w).Encode(response)
}
