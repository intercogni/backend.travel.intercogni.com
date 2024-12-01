package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

	http.HandleFunc("/api/bookings/create-complex", createComplexBooking)

	http.HandleFunc("/users/create", createUser)
	http.HandleFunc("/users/edit", editUser)
	http.HandleFunc("/users/delete", deleteUser)

	http.HandleFunc("/bookings/create", createBooking)
	http.HandleFunc("/bookings/edit", editBooking)
	http.HandleFunc("/bookings/delete", deleteBooking)

	http.HandleFunc("/bookings_persons/create", createBookingPerson)
	http.HandleFunc("/bookings_persons/edit", editBookingPerson)
	http.HandleFunc("/bookings_persons/delete", deleteBookingPerson)

	http.HandleFunc("/persons/create", createPerson)
	http.HandleFunc("/persons/edit", editPerson)
	http.HandleFunc("/persons/delete", deletePerson)

	http.HandleFunc("/vacations/create", createVacation)
	http.HandleFunc("/vacations/edit", editVacation)
	http.HandleFunc("/vacations/delete", deleteVacation)

	http.HandleFunc("/trip/create", createTrip)
	http.HandleFunc("/trip/edit", editTrip)
	http.HandleFunc("/trip/delete", deleteTrip)

	http.HandleFunc("/leg/create", createLeg)
	http.HandleFunc("/leg/edit", editLeg)
	http.HandleFunc("/leg/delete", deleteLeg)

	log.Fatal(http.ListenAndServe(":8080", nil))
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
	_, err = tx.Exec("INSERT INTO bookings (registrar_email, vacation_day_count, total_days, total_price) VALUES (?, ?, ?, ?)",
		booking.RegistrarEmail, booking.VacationDayCount, booking.TotalDays, booking.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	fmt.Println("testing vacation")
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
