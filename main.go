package main

import (
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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Booking struct {
	ID               uint     `gorm:"primaryKey"`
	RegistrarEmail   string   `gorm:"type:varchar(100);not null"`
	VacationDayCount float64  `gorm:"not null"`
	TotalPrice       float64  `gorm:"not null"`
	PricePerPax      float64  `gorm:"not null"`
	StartDate        string   `gorm:"type:varchar(20);not null"`
	EndDate          string   `gorm:"type:varchar(20);not null"`
	OutboundTripID   uint     `gorm:"not null"`
	InboundTripID    uint     `gorm:"not null"`
	VacationID       uint     `gorm:"not null"`
	Origin           string   `gorm:"type:varchar(100);not null"`
	Destination      string   `gorm:"type:varchar(100);not null"`
	People           []Person `gorm:"many2many:bookings_people;"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Person struct {
	ID             uint   `gorm:"primaryKey"`
	Nationality    string `gorm:"type:varchar(50);not null"`
	PassportNumber string `gorm:"type:varchar(50);not null"`
	FirstName      string `gorm:"type:varchar(50);not null"`
	LastName       string `gorm:"type:varchar(50);not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Trip struct {
	ID              uint    `gorm:"primaryKey"`
	DepartureFeeder uint    `gorm:"not null"`
	Trunk           uint    `gorm:"not null"`
	ArrivalFeeder   uint    `gorm:"not null"`
	TotalPrice      float64 `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Leg struct {
	ID              uint    `gorm:"primaryKey"`
	Type            string  `gorm:"type:varchar(50);not null"`
	Budget          string  `gorm:"type:varchar(50);not null"`
	OriginCity      string  `gorm:"type:varchar(100);not null"`
	DestinationCity string  `gorm:"type:varchar(100);not null"`
	Price           float64 `gorm:"not null"`
	ToBefore        float64 `gorm:"type:float"`
	ToNext          float64 `gorm:"type:float"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Vacation struct {
	ID                uint    `gorm:"primaryKey"`
	City              string  `gorm:"type:varchar(100);not null"`
	HotelBudget       string  `gorm:"type:varchar(50);not null"`
	SightseeingBudget string  `gorm:"type:varchar(50);not null"`
	TotalPrice        float64 `gorm:"not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("database.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Booking{}, &Person{}, &Trip{}, &Leg{}, &Vacation{})

	fmt.Printf("starting...\n")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/bookings/create-complex", createComplexBooking)
	mux.HandleFunc("/api/bookings/get-complex", getComplexBooking)
	mux.HandleFunc("/api/bookings/delete", deleteBooking)
	mux.HandleFunc("/api/set-general-info", updateGeneralInfo)
	mux.HandleFunc("/api/bookings/get-all", getAllBookings)

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

func deleteBooking(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BookingID uint `json:"booking_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.Delete(&Booking{}, request.BookingID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
	for _, airport := range airports {
		fmt.Printf("Checking airport: %s, %s\n", airport.City, airport.Name)
		distance := haversine(originLat, originLong, airport.Lat, airport.Long)
		if distance < minOriginDistance {
			minOriginDistance = distance
			closestOriginAirport = airport
		}
	}

	destinationLat := generalInfo.Destination.Lat
	destinationLong := generalInfo.Destination.Long
	var closestDestinationAirport Airport
	minDestinationDistance := math.MaxFloat64
	for _, airport := range airports {
		fmt.Printf("Checking airport: %s, %s\n", airport.City, airport.Name)
		distance := haversine(destinationLat, destinationLong, airport.Lat, airport.Long)
		if distance < minDestinationDistance {
			minDestinationDistance = distance
			closestDestinationAirport = airport
		}
	}

	originToAirportDist := haversine(generalInfo.Origin.Lat, generalInfo.Origin.Long, closestOriginAirport.Lat, closestOriginAirport.Long)
	airportToAirportDist := haversine(closestOriginAirport.Lat, closestOriginAirport.Long, closestDestinationAirport.Lat, closestDestinationAirport.Long)
	destinationToAirportDist := haversine(generalInfo.Destination.Lat, generalInfo.Destination.Long, closestDestinationAirport.Lat, closestDestinationAirport.Long)

	generalInfo.OriginAirport = struct {
		City     string  `json:"city"`
		Name     string  `json:"name"`
		Lat      float64 `json:"lat"`
		Long     float64 `json:"long"`
		IATACode string  `json:"iata_code"`
		ToBefore float64 `json:"to_before"`
		ToNext   float64 `json:"to_next"`
	}{
		City:     closestOriginAirport.City,
		Name:     closestOriginAirport.Name,
		Lat:      closestOriginAirport.Lat,
		Long:     closestOriginAirport.Long,
		IATACode: closestOriginAirport.IATACode,
		ToBefore: originToAirportDist,
		ToNext:   airportToAirportDist,
	}
	generalInfo.DestinationAirport = struct {
		City     string  `json:"city"`
		Name     string  `json:"name"`
		Lat      float64 `json:"lat"`
		Long     float64 `json:"long"`
		IATACode string  `json:"iata_code"`
		ToBefore float64 `json:"to_before"`
		ToNext   float64 `json:"to_next"`
	}{
		City:     closestDestinationAirport.City,
		Name:     closestDestinationAirport.Name,
		Lat:      closestDestinationAirport.Lat,
		Long:     closestDestinationAirport.Long,
		IATACode: closestDestinationAirport.IATACode,
		ToBefore: airportToAirportDist,
		ToNext:   destinationToAirportDist,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(generalInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getAllBookings(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bookings []Booking
	if err := db.Where("registrar_email = ?", request.Email).Find(&bookings).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bookings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createComplexBooking(w http.ResponseWriter, r *http.Request) {
	var booking struct {
		RegistrarEmail string `json:"registrar_email"`
		OutboundTrip   struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		} `json:"outbound_trip"`
		Vacation struct {
			City              string  `json:"city"`
			HotelBudget       string  `json:"hotel_budget"`
			SightseeingBudget string  `json:"sightseeing_budget"`
			TotalPrice        float64 `json:"total_price"`
		} `json:"vacation"`
		VacationDayCount float64 `json:"vacation_day_count"`
		InboundTrip      struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		} `json:"inbound_trip"`
		TotalDays   float64 `json:"total_days"`
		TotalPrice  float64 `json:"total_price"`
		PricePerPax float64 `json:"price_per_pax"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		Origin      string  `json:"origin"`
		Destination string  `json:"destination"`
		Persons     []struct {
			Nationality    string `json:"nationality"`
			PassportNumber string `json:"passport_number"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
		} `json:"persons"`
	}
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		log.Println("Error decoding booking:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var existingBooking Booking
	if err := tx.Where("registrar_email = ? AND vacation_day_count = ? AND total_price = ? AND start_date = ? AND end_date = ?",
		booking.RegistrarEmail, booking.VacationDayCount, booking.TotalPrice, booking.StartDate, booking.EndDate).First(&existingBooking).Error; err == nil {
		http.Error(w, "Booking already exists", http.StatusConflict)
		return
	}

	newBooking := Booking{
		RegistrarEmail:   booking.RegistrarEmail,
		VacationDayCount: booking.VacationDayCount,
		TotalPrice:       booking.TotalPrice,
		PricePerPax:      booking.PricePerPax,
		StartDate:        booking.StartDate,
		EndDate:          booking.EndDate,
		Origin:           booking.Origin,
		Destination:      booking.Destination,
	}

	if err := tx.Create(&newBooking).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("testing now\n")

	for _, person := range booking.Persons {
		var existingPerson Person
		if err := tx.Where("nationality = ? AND passport_number = ? AND first_name = ? AND last_name = ?",
			person.Nationality, person.PassportNumber, person.FirstName, person.LastName).First(&existingPerson).Error; err != nil {
			newPerson := Person{
				Nationality:    person.Nationality,
				PassportNumber: person.PassportNumber,
				FirstName:      person.FirstName,
				LastName:       person.LastName,
			}
			if err := tx.Create(&newPerson).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			existingPerson = newPerson
		}
		if err := tx.Model(&newBooking).Association("People").Append(&existingPerson); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Error appending person to booking:", err)
			return
		}
	}

	fmt.Printf("testing more\n")

	var vacation Vacation
	if err := tx.Where("city = ? AND hotel_budget = ? AND sightseeing_budget = ? AND total_price = ?",
		booking.Vacation.City, booking.Vacation.HotelBudget, booking.Vacation.SightseeingBudget, booking.Vacation.TotalPrice).First(&vacation).Error; err != nil {
		vacation = Vacation{
			City:              booking.Vacation.City,
			HotelBudget:       booking.Vacation.HotelBudget,
			SightseeingBudget: booking.Vacation.SightseeingBudget,
			TotalPrice:        booking.Vacation.TotalPrice,
		}
		if err := tx.Create(&vacation).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	createLeg := func(legData struct {
		Type            string  `json:"type"`
		Budget          string  `json:"budget"`
		OriginCity      string  `json:"origin_city"`
		DestinationCity string  `json:"destination_city"`
		Price           float64 `json:"price"`
		ToBefore        float64 `json:"to_before"`
		ToNext          float64 `json:"to_next"`
	}) (Leg, error) {
		var leg Leg
		if err := tx.Where("type = ? AND budget = ? AND origin_city = ? AND destination_city = ? AND price = ?",
			legData.Type, legData.Budget, legData.OriginCity, legData.DestinationCity, legData.Price).First(&leg).Error; err != nil {
			leg = Leg{
				Type:            legData.Type,
				Budget:          legData.Budget,
				OriginCity:      legData.OriginCity,
				DestinationCity: legData.DestinationCity,
				Price:           legData.Price,
				ToBefore:        legData.ToBefore,
				ToNext:          legData.ToNext,
			}
			if err := tx.Create(&leg).Error; err != nil {
				return leg, err
			}
		}
		return leg, nil
	}

	outboundDepartureFeeder, err := createLeg(booking.OutboundTrip.DepartureFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outboundTrunk, err := createLeg(booking.OutboundTrip.Trunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outboundArrivalFeeder, err := createLeg(booking.OutboundTrip.ArrivalFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inboundDepartureFeeder, err := createLeg(booking.InboundTrip.DepartureFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inboundTrunk, err := createLeg(booking.InboundTrip.Trunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inboundArrivalFeeder, err := createLeg(booking.InboundTrip.ArrivalFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	outboundTrip := Trip{
		DepartureFeeder: outboundDepartureFeeder.ID,
		Trunk:           outboundTrunk.ID,
		ArrivalFeeder:   outboundArrivalFeeder.ID,
		TotalPrice:      booking.OutboundTrip.TotalPrice,
	}
	if err := tx.Create(&outboundTrip).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inboundTrip := Trip{
		DepartureFeeder: inboundDepartureFeeder.ID,
		Trunk:           inboundTrunk.ID,
		ArrivalFeeder:   inboundArrivalFeeder.ID,
		TotalPrice:      booking.InboundTrip.TotalPrice,
	}
	if err := tx.Create(&inboundTrip).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newBooking.OutboundTripID = outboundTrip.ID
	newBooking.InboundTripID = inboundTrip.ID
	newBooking.VacationID = vacation.ID

	if err := tx.Save(&newBooking).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getComplexBooking(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BookingID uint `json:"booking_id"`
	}

	log.Printf("Request: %+v\n", request)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var booking Booking
	if err := db.Preload("People").First(&booking, request.BookingID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var outboundTrip Trip
	if err := db.First(&outboundTrip, booking.OutboundTripID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var inboundTrip Trip
	if err := db.First(&inboundTrip, booking.InboundTripID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var vacation Vacation
	if err := db.First(&vacation, booking.VacationID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getLeg := func(legID uint) (struct {
		Type            string  `json:"type"`
		Budget          string  `json:"budget"`
		OriginCity      string  `json:"origin_city"`
		DestinationCity string  `json:"destination_city"`
		Price           float64 `json:"price"`
		ToBefore        float64 `json:"to_before"`
		ToNext          float64 `json:"to_next"`
	}, error) {
		var leg Leg
		if err := db.First(&leg, legID).Error; err != nil {
			return struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			}{}, err
		}
		return struct {
			Type            string  `json:"type"`
			Budget          string  `json:"budget"`
			OriginCity      string  `json:"origin_city"`
			DestinationCity string  `json:"destination_city"`
			Price           float64 `json:"price"`
			ToBefore        float64 `json:"to_before"`
			ToNext          float64 `json:"to_next"`
		}{
			Type:            leg.Type,
			Budget:          leg.Budget,
			OriginCity:      leg.OriginCity,
			DestinationCity: leg.DestinationCity,
			Price:           leg.Price,
			ToBefore:        leg.ToBefore,
			ToNext:          leg.ToNext,
		}, nil
	}

	outboundDepartureFeeder, err := getLeg(outboundTrip.DepartureFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outboundTrunk, err := getLeg(outboundTrip.Trunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outboundArrivalFeeder, err := getLeg(outboundTrip.ArrivalFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inboundDepartureFeeder, err := getLeg(inboundTrip.DepartureFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inboundTrunk, err := getLeg(inboundTrip.Trunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inboundArrivalFeeder, err := getLeg(inboundTrip.ArrivalFeeder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		RegistrarEmail string `json:"registrar_email"`
		OutboundTrip   struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		} `json:"outbound_trip"`
		Vacation struct {
			City              string  `json:"city"`
			HotelBudget       string  `json:"hotel_budget"`
			SightseeingBudget string  `json:"sightseeing_budget"`
			TotalPrice        float64 `json:"total_price"`
		} `json:"vacation"`
		VacationDayCount float64 `json:"vacation_day_count"`
		InboundTrip      struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		} `json:"inbound_trip"`
		TotalDays   float64 `json:"total_days"`
		TotalPrice  float64 `json:"total_price"`
		PricePerPax float64 `json:"price_per_pax"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		Origin      string  `json:"origin"`
		Destination string  `json:"destination"`
		Persons     []struct {
			Nationality    string `json:"nationality"`
			PassportNumber string `json:"passport_number"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
		} `json:"persons"`
	}{
		RegistrarEmail: booking.RegistrarEmail,
		OutboundTrip: struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		}{
			DepartureFeeder: outboundDepartureFeeder,
			Trunk:           outboundTrunk,
			ArrivalFeeder:   outboundArrivalFeeder,
			TotalPrice:      outboundTrip.TotalPrice,
		},
		Vacation: struct {
			City              string  `json:"city"`
			HotelBudget       string  `json:"hotel_budget"`
			SightseeingBudget string  `json:"sightseeing_budget"`
			TotalPrice        float64 `json:"total_price"`
		}{
			City:              vacation.City,
			HotelBudget:       vacation.HotelBudget,
			SightseeingBudget: vacation.SightseeingBudget,
			TotalPrice:        vacation.TotalPrice,
		},
		VacationDayCount: booking.VacationDayCount,
		InboundTrip: struct {
			DepartureFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"departure_feeder"`
			Trunk struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"trunk"`
			ArrivalFeeder struct {
				Type            string  `json:"type"`
				Budget          string  `json:"budget"`
				OriginCity      string  `json:"origin_city"`
				DestinationCity string  `json:"destination_city"`
				Price           float64 `json:"price"`
				ToBefore        float64 `json:"to_before"`
				ToNext          float64 `json:"to_next"`
			} `json:"arrival_feeder"`
			TotalPrice float64 `json:"total_price"`
		}{
			DepartureFeeder: inboundDepartureFeeder,
			Trunk:           inboundTrunk,
			ArrivalFeeder:   inboundArrivalFeeder,
			TotalPrice:      inboundTrip.TotalPrice,
		},
		TotalDays:   booking.VacationDayCount,
		TotalPrice:  booking.TotalPrice,
		PricePerPax: booking.PricePerPax,
		StartDate:   booking.StartDate,
		EndDate:     booking.EndDate,
		Origin:      booking.Origin,
		Destination: booking.Destination,
		Persons: func() []struct {
			Nationality    string `json:"nationality"`
			PassportNumber string `json:"passport_number"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
		} {
			var persons []struct {
				Nationality    string `json:"nationality"`
				PassportNumber string `json:"passport_number"`
				FirstName      string `json:"first_name"`
				LastName       string `json:"last_name"`
			}
			for _, person := range booking.People {
				persons = append(persons, struct {
					Nationality    string `json:"nationality"`
					PassportNumber string `json:"passport_number"`
					FirstName      string `json:"first_name"`
					LastName       string `json:"last_name"`
				}{
					Nationality:    person.Nationality,
					PassportNumber: person.PassportNumber,
					FirstName:      person.FirstName,
					LastName:       person.LastName,
				})
			}
			return persons
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
