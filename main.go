package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "database.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Printf("starting...\n")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/bookings/create-complex", createComplexBooking)
	mux.HandleFunc("/api/set-general-info", updateGeneralInfo)
	mux.HandleFunc("/users/create", createUser)
	mux.HandleFunc("/users/edit", editUser)
	mux.HandleFunc("/users/delete", deleteUser)
	mux.HandleFunc("/bookings/create", createBooking)
	mux.HandleFunc("/bookings/edit", editBooking)
	mux.HandleFunc("/bookings/delete", deleteBooking)
	mux.HandleFunc("/bookings_persons/create", createBookingPerson)
	mux.HandleFunc("/bookings_persons/edit", editBookingPerson)
	mux.HandleFunc("/bookings_persons/delete", deleteBookingPerson)
	mux.HandleFunc("/persons/create", createPerson)
	mux.HandleFunc("/persons/edit", editPerson)
	mux.HandleFunc("/persons/delete", deletePerson)
	mux.HandleFunc("/vacations/create", createVacation)
	mux.HandleFunc("/vacations/edit", editVacation)
	mux.HandleFunc("/vacations/delete", deleteVacation)
	mux.HandleFunc("/trip/create", createTrip)
	mux.HandleFunc("/trip/edit", editTrip)
	mux.HandleFunc("/trip/delete", deleteTrip)
	mux.HandleFunc("/leg/create", createLeg)
	mux.HandleFunc("/leg/edit", editLeg)
	mux.HandleFunc("/leg/delete", deleteLeg)

	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
}

type Airport struct {
	City     string
	Name     string
	Lat      float64
	Long     float64
	IATACode string
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in kilometers
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func loadAirports(filePath string) ([]Airport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var airports []Airport
	for _, record := range records[1:] {
		lat, _ := strconv.ParseFloat(record[4], 64)
		long, _ := strconv.ParseFloat(record[5], 64)
		// fmt.Printf("Lat: %f, Long: %f\n", lat, long)
		airports = append(airports, Airport{
			City:     record[10],
			Name:     record[3],
			Lat:      lat,
			Long:     long,
			IATACode: record[13],
		})
	}
	return airports, nil
}

// func updateGeneralInfo(w http.ResponseWriter, r *http.Request) {
// 	var generalInfo struct {
// 		OriginCity struct {
// 			Name string `json:"name"`
// 			Lat  string `json:"lat"`
// 			Long string `json:"long"`
// 		} `json:"origin_city"`
// 		OriginAirportCity struct {
// 			Name        string `json:"name"`
// 			Description string `json:"description"`
// 			Lat         string `json:"lat"`
// 			Long        string `json:"long"`
// 			IATACode    string `json:"iata_code"`
// 		} `json:"origin_airport_city"`
// 		DestinationAirportCity struct {
// 			Name        string `json:"name"`
// 			Description string `json:"description"`
// 			Lat         string `json:"lat"`
// 			Long        string `json:"long"`
// 			IATACode    string `json:"iata_code"`
// 		} `json:"destination_airport_city"`
// 		DestinationCity struct {
// 			Name string `json:"name"`
// 			Lat  string `json:"lat"`
// 			Long string `json:"long"`
// 		} `json:"destination_city"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&generalInfo); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	fmt.Printf("%+v\n", generalInfo)

// airports, err := loadAirports("large_airports.csv")
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }

// originLat, _ := strconv.ParseFloat(generalInfo.OriginCity.Lat, 64)
// originLong, _ := strconv.ParseFloat(generalInfo.OriginCity.Long, 64)

// var closestAirport Airport
// minDistance := math.MaxFloat64

// for _, airport := range airports {
// 	distance := haversine(originLat, originLong, airport.Lat, airport.Long)
// 	if distance < minDistance {
// 		minDistance = distance
// 		closestAirport = airport
// 	}
// }

// generalInfo.OriginAirportCity = struct {
// 	Name        string `json:"name"`
// 	Description string `json:"description"`
// 	Lat         string `json:"lat"`
// 	Long        string `json:"long"`
// 	IATACode    string `json:"iata_code"`
// }{
// 	Name:        closestAirport.Name,
// 	Description: closestAirport.Description,
// 	Lat:         strconv.FormatFloat(closestAirport.Lat, 'f', 6, 64),
// 	Long:        strconv.FormatFloat(closestAirport.Long, 'f', 6, 64),
// 	IATACode:    closestAirport.IATACode,
// }

// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(generalInfo); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

func updateGeneralInfo(w http.ResponseWriter, r *http.Request) {
	var generalInfo struct {
		Origin struct {
			Country string  `json:"country"`
			State   string  `json:"state"`
			City    string  `json:"city"`
			Lat     float64 `json:"lat"`
			Long    float64 `json:"long"`
			ToNext  float64 `json:"to_next"`
		} `json:"origin"`
		OriginAirport struct {
			City     string  `json:"city"`
			Name     string  `json:"name"`
			Lat      float64 `json:"lat"`
			Long     float64 `json:"long"`
			IATACode string  `json:"iata_code"`
			ToBefore float64 `json:"to_before"`
			ToNext   float64 `json:"to_next"`
		} `json:"origin_airport"`
		DestinationAirport struct {
			City     string  `json:"city"`
			Name     string  `json:"name"`
			Lat      float64 `json:"lat"`
			Long     float64 `json:"long"`
			IATACode string  `json:"iata_code"`
			ToBefore float64 `json:"to_before"`
			ToNext   float64 `json:"to_next"`
		} `json:"destination_airport"`
		Destination struct {
			Country  string  `json:"country"`
			State    string  `json:"state"`
			City     string  `json:"city"`
			Lat      float64 `json:"lat"`
			Long     float64 `json:"long"`
			ToBefore float64 `json:"to_before"`
		} `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&generalInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v\n", generalInfo)

	generalInfo.OriginAirport.Name = "lorem ipsum"
	generalInfo.DestinationAirport.Name = "lorem ipsum"

	airports, err := loadAirports("large_airports.csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	originLat := generalInfo.Origin.Lat
	originLong := generalInfo.Origin.Long
	var closestOriginAirport Airport
	minOriginDistance := math.MaxFloat64
	// for _, airport := range airports {
	// 	fmt.Printf("City: %s, Name: %s, Lat: %f, Long: %f, IATACode: %s\n", airport.City, airport.Name, airport.Lat, airport.Long, airport.IATACode)
	// }
	for _, airport := range airports {
		fmt.Printf("Checking airport: %s, %s\n", airport.City, airport.Name)
		distance := haversine(originLat, originLong, airport.Lat, airport.Long)
		if distance < minOriginDistance {
			minOriginDistance = distance
			closestOriginAirport = airport
		}
	}
	generalInfo.OriginAirport = struct {
		City     string  `json:"city"`
		Name     string  `json:"name"`
		Lat      float64 `json:"lat"`
		Long     float64 `json:"long"`
		IATACode string  `json:"iata_code"`
	}{
		City:     closestOriginAirport.City,
		Name:     closestOriginAirport.Name,
		Lat:      closestOriginAirport.Lat,
		Long:     closestOriginAirport.Long,
		IATACode: closestOriginAirport.IATACode,
	}

	destinationLat := generalInfo.Destination.Lat
	destinationLong := generalInfo.Destination.Long
	var closestDestinationAirport Airport
	minDestinationDistance := math.MaxFloat64
	// for _, airport := range airports {
	// 	fmt.Printf("City: %s, Name: %s, Lat: %f, Long: %f, IATACode: %s\n", airport.City, airport.Name, airport.Lat, airport.Long, airport.IATACode)
	// }
	for _, airport := range airports {
		fmt.Printf("Checking airport: %s, %s\n", airport.City, airport.Name)
		distance := haversine(destinationLat, destinationLong, airport.Lat, airport.Long)
		if distance < minDestinationDistance {
			minDestinationDistance = distance
			closestDestinationAirport = airport
		}
	}
	generalInfo.DestinationAirport = struct {
		City     string  `json:"city"`
		Name     string  `json:"name"`
		Lat      float64 `json:"lat"`
		Long     float64 `json:"long"`
		IATACode string  `json:"iata_code"`
	}{
		City:     closestDestinationAirport.City,
		Name:     closestDestinationAirport.Name,
		Lat:      closestDestinationAirport.Lat,
		Long:     closestDestinationAirport.Long,
		IATACode: closestDestinationAirport.IATACode,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(generalInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		GithubEmail  string    `json:"github_email"`
		Name         string    `json:"name"`
		RegisteredAt time.Time `json:"registered_at"`
		LastLogin    time.Time `json:"last_login"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO users (github_email, name, registered_at, last_login) VALUES (?, ?, ?, ?)",
		user.GithubEmail, user.Name, user.RegisteredAt, user.LastLogin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		GithubEmail  string    `json:"github_email"`
		Name         string    `json:"name"`
		RegisteredAt time.Time `json:"registered_at"`
		LastLogin    time.Time `json:"last_login"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE users SET name = ?, registered_at = ?, last_login = ? WHERE github_email = ?",
		user.Name, user.RegisteredAt, user.LastLogin, user.GithubEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		GithubEmail string `json:"github_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM users WHERE github_email = ?", user.GithubEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createComplexBooking(w http.ResponseWriter, r *http.Request) {
	var booking struct {
		RegistrarEmail string `json:"registrar_email"`
		OutboundTrip   struct {
			DepartureFeeder struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"arrival_feeder"`
			TotalPrice int `json:"total_price"`
		} `json:"outbound_trip"`
		Vacation struct {
			City              string `json:"city"`
			HotelBudget       string `json:"hotel_budget"`
			SightseeingBudget string `json:"sightseeing_budget"`
			TotalPrice        int    `json:"total_price"`
		} `json:"vacation"`
		VacationDayCount int `json:"vacation_day_count"`
		InboundTrip      struct {
			DepartureFeeder struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string `json:"type"`
				Budget          string `json:"budget"`
				OriginCity      string `json:"origin_city"`
				DestinationCity string `json:"destination_city"`
				Price           int    `json:"price"`
			} `json:"arrival_feeder"`
			TotalPrice int `json:"total_price"`
		} `json:"inbound_trip"`
		TotalDays  int `json:"total_days"`
		TotalPrice int `json:"total_price"`
		Persons    []struct {
			Nationality    string `json:"nationality"`
			PassportNumber string `json:"passport_number"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
		} `json:"persons"`
	}
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var exists bool

	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM bookings WHERE registrar_email = ? AND vacation_day_count = ? AND total_days = ? AND total_price = ?)",
		booking.RegistrarEmail, booking.VacationDayCount, booking.TotalDays, booking.TotalPrice).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Booking already exists", http.StatusConflict)
		return
	}

	fmt.Println("Inserting booking")
	res, err := tx.Exec("INSERT INTO bookings (registrar_email, vacation_day_count, total_days, total_price) VALUES (?, ?, ?, ?)",
		booking.RegistrarEmail, booking.VacationDayCount, booking.TotalDays, booking.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bookingID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert persons and link them to the booking
	for _, person := range booking.Persons {
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM persons WHERE nationality = ? AND passport_number = ? AND first_name = ? AND last_name = ?)",
			person.Nationality, person.PassportNumber, person.FirstName, person.LastName).Scan(&exists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			fmt.Println("Inserting person")
			_, err = tx.Exec("INSERT INTO persons (nationality, passport_number, first_name, last_name) VALUES (?, ?, ?, ?)",
				person.Nationality, person.PassportNumber, person.FirstName, person.LastName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		var personID int
		err = tx.QueryRow("SELECT id FROM persons WHERE nationality = ? AND passport_number = ? AND first_name = ? AND last_name = ?",
			person.Nationality, person.PassportNumber, person.FirstName, person.LastName).Scan(&personID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("Linking person to booking")
		_, err = tx.Exec("INSERT INTO bookings_persons (booking_id, person_id) VALUES (?, ?)", bookingID, personID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Insert vacation
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM vacations WHERE city = ? AND hotel_budget = ? AND sightseeing_budget = ? AND total_price = ?)",
		booking.Vacation.City, booking.Vacation.HotelBudget, booking.Vacation.SightseeingBudget, booking.Vacation.TotalPrice).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting vacation")
		_, err = tx.Exec("INSERT INTO vacations (city, hotel_budget, sightseeing_budget, total_price) VALUES (?, ?, ?, ?)",
			booking.Vacation.City, booking.Vacation.HotelBudget, booking.Vacation.SightseeingBudget, booking.Vacation.TotalPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing outbound trip departure feeder")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.OutboundTrip.DepartureFeeder.Type, booking.OutboundTrip.DepartureFeeder.Budget, booking.OutboundTrip.DepartureFeeder.OriginCity, booking.OutboundTrip.DepartureFeeder.DestinationCity, booking.OutboundTrip.DepartureFeeder.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting outbound trip departure feeder")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.OutboundTrip.DepartureFeeder.Type, booking.OutboundTrip.DepartureFeeder.Budget, booking.OutboundTrip.DepartureFeeder.OriginCity, booking.OutboundTrip.DepartureFeeder.DestinationCity, booking.OutboundTrip.DepartureFeeder.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing outbound trip trunk")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.OutboundTrip.Trunk.Type, booking.OutboundTrip.Trunk.Budget, booking.OutboundTrip.Trunk.OriginCity, booking.OutboundTrip.Trunk.DestinationCity, booking.OutboundTrip.Trunk.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting outbound trip trunk")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.OutboundTrip.Trunk.Type, booking.OutboundTrip.Trunk.Budget, booking.OutboundTrip.Trunk.OriginCity, booking.OutboundTrip.Trunk.DestinationCity, booking.OutboundTrip.Trunk.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing outbound trip arrival feeder")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.OutboundTrip.ArrivalFeeder.Type, booking.OutboundTrip.ArrivalFeeder.Budget, booking.OutboundTrip.ArrivalFeeder.OriginCity, booking.OutboundTrip.ArrivalFeeder.DestinationCity, booking.OutboundTrip.ArrivalFeeder.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting outbound trip arrival feeder")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.OutboundTrip.ArrivalFeeder.Type, booking.OutboundTrip.ArrivalFeeder.Budget, booking.OutboundTrip.ArrivalFeeder.OriginCity, booking.OutboundTrip.ArrivalFeeder.DestinationCity, booking.OutboundTrip.ArrivalFeeder.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing inbound trip departure feeder")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.InboundTrip.DepartureFeeder.Type, booking.InboundTrip.DepartureFeeder.Budget, booking.InboundTrip.DepartureFeeder.OriginCity, booking.InboundTrip.DepartureFeeder.DestinationCity, booking.InboundTrip.DepartureFeeder.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting inbound trip departure feeder")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.InboundTrip.DepartureFeeder.Type, booking.InboundTrip.DepartureFeeder.Budget, booking.InboundTrip.DepartureFeeder.OriginCity, booking.InboundTrip.DepartureFeeder.DestinationCity, booking.InboundTrip.DepartureFeeder.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing inbound trip trunk")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.InboundTrip.Trunk.Type, booking.InboundTrip.Trunk.Budget, booking.InboundTrip.Trunk.OriginCity, booking.InboundTrip.Trunk.DestinationCity, booking.InboundTrip.Trunk.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting inbound trip trunk")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.InboundTrip.Trunk.Type, booking.InboundTrip.Trunk.Budget, booking.InboundTrip.Trunk.OriginCity, booking.InboundTrip.Trunk.DestinationCity, booking.InboundTrip.Trunk.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("testing inbound trip arrival feeder")
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		booking.InboundTrip.ArrivalFeeder.Type, booking.InboundTrip.ArrivalFeeder.Budget, booking.InboundTrip.ArrivalFeeder.OriginCity, booking.InboundTrip.ArrivalFeeder.DestinationCity, booking.InboundTrip.ArrivalFeeder.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("Inserting inbound leg arrival feeder")
		_, err = tx.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
			booking.InboundTrip.ArrivalFeeder.Type, booking.InboundTrip.ArrivalFeeder.Budget, booking.InboundTrip.ArrivalFeeder.OriginCity, booking.InboundTrip.ArrivalFeeder.DestinationCity, booking.InboundTrip.ArrivalFeeder.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("inserting outbound legs into trip")
	var outboundDepartureFeederID, outboundTrunkID, outboundArrivalFeederID int

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.OutboundTrip.DepartureFeeder.Type, booking.OutboundTrip.DepartureFeeder.Budget, booking.OutboundTrip.DepartureFeeder.OriginCity, booking.OutboundTrip.DepartureFeeder.DestinationCity, booking.OutboundTrip.DepartureFeeder.Price).Scan(&outboundDepartureFeederID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.OutboundTrip.Trunk.Type, booking.OutboundTrip.Trunk.Budget, booking.OutboundTrip.Trunk.OriginCity, booking.OutboundTrip.Trunk.DestinationCity, booking.OutboundTrip.Trunk.Price).Scan(&outboundTrunkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.OutboundTrip.ArrivalFeeder.Type, booking.OutboundTrip.ArrivalFeeder.Budget, booking.OutboundTrip.ArrivalFeeder.OriginCity, booking.OutboundTrip.ArrivalFeeder.DestinationCity, booking.OutboundTrip.ArrivalFeeder.Price).Scan(&outboundArrivalFeederID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Inserting outbound trip into trip table")
	_, err = tx.Exec("INSERT INTO trip (departure_feeder, trunk, arrival_feeder, total_price) VALUES (?, ?, ?, ?)",
		outboundDepartureFeederID, outboundTrunkID, outboundArrivalFeederID, booking.OutboundTrip.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("inserting inbound legs into trip")
	var inboundDepartureFeederID, inboundTrunkID, inboundArrivalFeederID int

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.InboundTrip.DepartureFeeder.Type, booking.InboundTrip.DepartureFeeder.Budget, booking.InboundTrip.DepartureFeeder.OriginCity, booking.InboundTrip.DepartureFeeder.DestinationCity, booking.InboundTrip.DepartureFeeder.Price).Scan(&inboundDepartureFeederID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.InboundTrip.Trunk.Type, booking.InboundTrip.Trunk.Budget, booking.InboundTrip.Trunk.OriginCity, booking.InboundTrip.Trunk.DestinationCity, booking.InboundTrip.Trunk.Price).Scan(&inboundTrunkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.QueryRow("SELECT id FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
		booking.InboundTrip.ArrivalFeeder.Type, booking.InboundTrip.ArrivalFeeder.Budget, booking.InboundTrip.ArrivalFeeder.OriginCity, booking.InboundTrip.ArrivalFeeder.DestinationCity, booking.InboundTrip.ArrivalFeeder.Price).Scan(&inboundArrivalFeederID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Inserting inbound trip into trip table")
	_, err = tx.Exec("INSERT INTO trip (departure_feeder, trunk, arrival_feeder, total_price) VALUES (?, ?, ?, ?)",
		inboundDepartureFeederID, inboundTrunkID, inboundArrivalFeederID, booking.InboundTrip.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var outboundTripID int
	err = tx.QueryRow("SELECT id FROM trip WHERE departure_feeder = ? AND trunk = ? AND arrival_feeder = ? AND total_price = ?",
		outboundDepartureFeederID, outboundTrunkID, outboundArrivalFeederID, booking.OutboundTrip.TotalPrice).Scan(&outboundTripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var inboundTripID int
	err = tx.QueryRow("SELECT id FROM trip WHERE departure_feeder = ? AND trunk = ? AND arrival_feeder = ? AND total_price = ?",
		inboundDepartureFeederID, inboundTrunkID, inboundArrivalFeederID, booking.InboundTrip.TotalPrice).Scan(&inboundTripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var vacationID int
	err = tx.QueryRow("SELECT id FROM vacations WHERE city = ? AND hotel_budget = ? AND sightseeing_budget = ? AND total_price = ?",
		booking.Vacation.City, booking.Vacation.HotelBudget, booking.Vacation.SightseeingBudget, booking.Vacation.TotalPrice).Scan(&vacationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE bookings SET registrar_email = ?, outbound_trip = ?, vacation = ?, vacation_day_count = ?, inbound_trip = ?, total_days = ?, total_price = ? WHERE id = ?",
		booking.RegistrarEmail, outboundTripID, vacationID, booking.VacationDayCount, inboundTripID, booking.TotalDays, booking.TotalPrice, bookingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func createBooking(w http.ResponseWriter, r *http.Request) {
	var booking struct {
		RegistrarEmail   string `json:"registrar_email"`
		OutboundTrip     int    `json:"outbound_trip"`
		Vacation         int    `json:"vacation"`
		VacationDayCount int    `json:"vacation_day_count"`
		InboundTrip      int    `json:"inbound_trip"`
		TotalDays        int    `json:"total_days"`
		TotalPrice       int    `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM bookings WHERE registrar_email = ? AND outbound_trip = ? AND vacation = ? AND vacation_day_count = ? AND inbound_trip = ? AND total_days = ? AND total_price = ?)",
		booking.RegistrarEmail, booking.OutboundTrip, booking.Vacation, booking.VacationDayCount, booking.InboundTrip, booking.TotalDays, booking.TotalPrice).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Booking already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO bookings (registrar_email, outbound_trip, vacation, vacation_day_count, inbound_trip, total_days, total_price) VALUES (?, ?, ?, ?, ?, ?, ?)",
		booking.RegistrarEmail, booking.OutboundTrip, booking.Vacation, booking.VacationDayCount, booking.InboundTrip, booking.TotalDays, booking.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editBooking(w http.ResponseWriter, r *http.Request) {
	var booking struct {
		ID               int    `json:"id"`
		RegistrarEmail   string `json:"registrar_email"`
		OutboundTrip     int    `json:"outbound_trip"`
		Vacation         int    `json:"vacation"`
		VacationDayCount int    `json:"vacation_day_count"`
		InboundTrip      int    `json:"inbound_trip"`
		TotalDays        int    `json:"total_days"`
		TotalPrice       int    `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE bookings SET registrar_email = ?, outbound_trip = ?, vacation = ?, vacation_day_count = ?, inbound_trip = ?, total_days = ?, total_price = ? WHERE id = ?",
		booking.RegistrarEmail, booking.OutboundTrip, booking.Vacation, booking.VacationDayCount, booking.InboundTrip, booking.TotalDays, booking.TotalPrice, booking.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteBooking(w http.ResponseWriter, r *http.Request) {
	var booking struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM bookings WHERE id = ?", booking.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createBookingPerson(w http.ResponseWriter, r *http.Request) {
	var bookingPerson struct {
		BookingID int `json:"booking_id"`
		PersonID  int `json:"person_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&bookingPerson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM bookings_persons WHERE booking_id = ? AND person_id = ?)",
		bookingPerson.BookingID, bookingPerson.PersonID).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Booking person already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO bookings_persons (booking_id, person_id) VALUES (?, ?)",
		bookingPerson.BookingID, bookingPerson.PersonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editBookingPerson(w http.ResponseWriter, r *http.Request) {
	var bookingPerson struct {
		BookingID int `json:"booking_id"`
		PersonID  int `json:"person_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&bookingPerson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE bookings_persons SET person_id = ? WHERE booking_id = ?",
		bookingPerson.PersonID, bookingPerson.BookingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteBookingPerson(w http.ResponseWriter, r *http.Request) {
	var bookingPerson struct {
		BookingID int `json:"booking_id"`
		PersonID  int `json:"person_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&bookingPerson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM bookings_persons WHERE booking_id = ? AND person_id = ?",
		bookingPerson.BookingID, bookingPerson.PersonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	var person struct {
		Nationality    string `json:"nationality"`
		PassportNumber string `json:"passport_number"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM persons WHERE nationality = ? AND passport_number = ? AND first_name = ? AND last_name = ?)",
		person.Nationality, person.PassportNumber, person.FirstName, person.LastName).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Person already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO persons (nationality, passport_number, first_name, last_name) VALUES (?, ?, ?, ?)",
		person.Nationality, person.PassportNumber, person.FirstName, person.LastName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editPerson(w http.ResponseWriter, r *http.Request) {
	var person struct {
		ID             int    `json:"id"`
		Nationality    string `json:"nationality"`
		PassportNumber string `json:"passport_number"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE persons SET nationality = ?, passport_number = ?, first_name = ?, last_name = ? WHERE id = ?",
		person.Nationality, person.PassportNumber, person.FirstName, person.LastName, person.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	var person struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM persons WHERE id = ?", person.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createVacation(w http.ResponseWriter, r *http.Request) {
	var vacation struct {
		City              string `json:"city"`
		HotelBudget       string `json:"hotel_budget"`
		SightseeingBudget string `json:"sightseeing_budget"`
		TotalPrice        int    `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&vacation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM vacations WHERE city = ? AND hotel_budget = ? AND sightseeing_budget = ? AND total_price = ?)",
		vacation.City, vacation.HotelBudget, vacation.SightseeingBudget, vacation.TotalPrice).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Vacation already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO vacations (city, hotel_budget, sightseeing_budget, total_price) VALUES (?, ?, ?, ?)",
		vacation.City, vacation.HotelBudget, vacation.SightseeingBudget, vacation.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editVacation(w http.ResponseWriter, r *http.Request) {
	var vacation struct {
		ID                int    `json:"id"`
		City              string `json:"city"`
		HotelBudget       string `json:"hotel_budget"`
		SightseeingBudget string `json:"sightseeing_budget"`
		TotalPrice        int    `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&vacation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE vacations SET city = ?, hotel_budget = ?, sightseeing_budget = ?, total_price = ? WHERE id = ?",
		vacation.City, vacation.HotelBudget, vacation.SightseeingBudget, vacation.TotalPrice, vacation.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteVacation(w http.ResponseWriter, r *http.Request) {
	var vacation struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&vacation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM vacations WHERE id = ?", vacation.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createTrip(w http.ResponseWriter, r *http.Request) {
	var trip struct {
		DepartureFeeder int `json:"departure_feeder"`
		Trunk           int `json:"trunk"`
		ArrivalFeeder   int `json:"arrival_feeder"`
		TotalPrice      int `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&trip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM trip WHERE departure_feeder = ? AND trunk = ? AND arrival_feeder = ? AND total_price = ?)",
		trip.DepartureFeeder, trip.Trunk, trip.ArrivalFeeder, trip.TotalPrice).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Trip already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO trip (departure_feeder, trunk, arrival_feeder, total_price) VALUES (?, ?, ?, ?)",
		trip.DepartureFeeder, trip.Trunk, trip.ArrivalFeeder, trip.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editTrip(w http.ResponseWriter, r *http.Request) {
	var trip struct {
		ID              int `json:"id"`
		DepartureFeeder int `json:"departure_feeder"`
		Trunk           int `json:"trunk"`
		ArrivalFeeder   int `json:"arrival_feeder"`
		TotalPrice      int `json:"total_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&trip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE trip SET departure_feeder = ?, trunk = ?, arrival_feeder = ?, total_price = ? WHERE id = ?",
		trip.DepartureFeeder, trip.Trunk, trip.ArrivalFeeder, trip.TotalPrice, trip.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteTrip(w http.ResponseWriter, r *http.Request) {
	var trip struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&trip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM trip WHERE id = ?", trip.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createLeg(w http.ResponseWriter, r *http.Request) {
	var leg struct {
		Type            string `json:"type"`
		Budget          string `json:"budget"`
		OriginCity      string `json:"origin_city"`
		DestinationCity string `json:"destination_city"`
		Price           int    `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&leg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM leg WHERE type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?)",
		leg.Type, leg.Budget, leg.OriginCity, leg.DestinationCity, leg.Price).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Leg already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO leg (type, budget, origin_city, destination_city, price) VALUES (?, ?, ?, ?, ?)",
		leg.Type, leg.Budget, leg.OriginCity, leg.DestinationCity, leg.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editLeg(w http.ResponseWriter, r *http.Request) {
	var leg struct {
		ID              int    `json:"id"`
		Type            string `json:"type"`
		Budget          string `json:"budget"`
		OriginCity      string `json:"origin_city"`
		DestinationCity string `json:"destination_city"`
		Price           int    `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&leg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE leg SET type = ?, budget = ?, origin_city = ?, destination_city = ?, price = ? WHERE id = ?",
		leg.Type, leg.Budget, leg.OriginCity, leg.DestinationCity, leg.Price, leg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteLeg(w http.ResponseWriter, r *http.Request) {
	var leg struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&leg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM leg WHERE id = ?", leg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
