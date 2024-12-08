package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Promotion struct
type Promotion struct {
	PromoCode         string  `json:"promo_code"`
	PromotionName     string  `json:"promotion_name"`
	PromotionDiscount float64 `json:"discount_percentage"`
	ValidFrom         string  `json:"valid_from"`
	ValidTo           string  `json:"valid_to"`
}

var db *sql.DB

/* Initialise the user_svc_db database connection
func initDB() {
	// Load environment variables from the .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}

	// Retrieve the DB host and port from environment variables
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")

	// Ensure that DB_HOST and DB_PORT are not empty
	if DB_HOST == "" || DB_PORT == "" {
		log.Fatal("DB_HOST and DB_PORT must be set")
		return
	}

	// Format the connection string using user credentials, host, port, and database name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		"user",
		"password",
		DB_HOST,
		DB_PORT,
		"promotion_svc_db",
	)

	// Open a connection to the database
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
		return
	}
}
*/

func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(host.docker.internal:3306)/promotion_svc_db")
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
	router.HandleFunc("/api/v1/promotions", getAllPromotions).Methods("GET")
	router.HandleFunc("/api/v1/promotions/{promo_code}", getPromotionByPromoCode).Methods("GET")
	fmt.Println("Listening at port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Creating a fuction to get promotion details by promotion_code
func getAllPromotions(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message    string      `json:"message"`
		Promotions []Promotion `json:"promotions"`
	}

	var promotions []Promotion

	// Query to get all promotions
	query := "SELECT promo_code, promotion_name, discount_percentage, valid_from, valid_to FROM promotion"

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		// If there is an error
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Internal server error", nil}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	// Loop through the rows and append promotions to the slice
	for rows.Next() {
		var promotion Promotion
		if err := rows.Scan(&promotion.PromoCode, &promotion.PromotionName, &promotion.PromotionDiscount, &promotion.ValidFrom, &promotion.ValidTo); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := Response{"Error scanning promotions", nil}
			json.NewEncoder(w).Encode(response)
			return
		}
		promotions = append(promotions, promotion)
	}

	// Check for any error encountered during the rows iteration
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := Response{"Error iterating promotions", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// If promotions found
	if len(promotions) > 0 {
		w.WriteHeader(http.StatusOK)
		response := Response{"Promotions found", promotions}
		json.NewEncoder(w).Encode(response)
	} else {
		// If no promotions found
		w.WriteHeader(http.StatusNotFound)
		response := Response{"No promotions found", nil}
		json.NewEncoder(w).Encode(response)
	}
}

// Creating a fuction to get promotion details by promotion_code
func getPromotionByPromoCode(w http.ResponseWriter, r *http.Request) {
	// Set the response header
	w.Header().Set("Content-Type", "application/json")

	// Struct for response
	type Response struct {
		Message   string     `json:"message"`
		Promotion *Promotion `json:"promotion"`
	}

	// Get the promotion_code from the request
	params := mux.Vars(r)
	promoCode := params["promo_code"]

	// Query to get promotion details by promotion_code
	query := "SELECT promo_code, promotion_name, discount_percentage, valid_from, valid_to FROM promotion WHERE promo_code = ?"

	// Execute the query
	row := db.QueryRow(query, promoCode)

	// Scan the row and assign to the Promotion struct
	var promotion Promotion
	if err := row.Scan(&promotion.PromoCode, &promotion.PromotionName, &promotion.PromotionDiscount, &promotion.ValidFrom, &promotion.ValidTo); err != nil {
		// If there is an error
		w.WriteHeader(http.StatusNotFound)
		response := Response{"Promotion not found", nil}
		json.NewEncoder(w).Encode(response)
		return
	}

	// If promotion found
	w.WriteHeader(http.StatusOK)
	response := Response{"Promotion found", &promotion}
	json.NewEncoder(w).Encode(response)
}
