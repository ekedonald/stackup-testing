package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Driver struct {
	ID           string
	Name         string
	Location     [2]float64
	LocationName string
	Distance     float64
	Deliveries   int
}

var drivers = []Driver{
	{ID: "driver1", Name: "John", Location: [2]float64{40.730610, -73.935242}, Distance: 0, Deliveries: 0},
	{ID: "driver2", Name: "Jane", Location: [2]float64{34.052235, -118.243683}, Distance: 0, Deliveries: 0},
	{ID: "driver3", Name: "Alex", Location: [2]float64{37.774929, -122.419418}, Distance: 0, Deliveries: 0},
	{ID: "driver4", Name: "Emma", Location: [2]float64{41.878113, -87.629799}, Distance: 0, Deliveries: 0},
	{ID: "driver5", Name: "Michael", Location: [2]float64{51.507351, -0.127758}, Distance: 0, Deliveries: 0},
	{ID: "driver6", Name: "Sophia", Location: [2]float64{48.856614, 2.352222}, Distance: 0, Deliveries: 0},
	{ID: "driver7", Name: "David", Location: [2]float64{35.689487, 139.691706}, Distance: 0, Deliveries: 0},
	{ID: "driver8", Name: "Olivia", Location: [2]float64{-33.868820, 151.209296}, Distance: 0, Deliveries: 0},
}

func getLocationName(lat, lon float64) string {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching location name:", err)
		return "Unknown"
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return "Unknown"
	}

	address, ok := result["address"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	street := address["road"]
	city := address["city"]
	state := address["state"]
	country := address["country"]

	var locationParts []string
	if street != nil {
		locationParts = append(locationParts, street.(string))
	}
	if city != nil {
		locationParts = append(locationParts, city.(string))
	}
	if state != nil {
		locationParts = append(locationParts, state.(string))
	}
	if country != nil {
		locationParts = append(locationParts, country.(string))
	}

	return strings.Join(locationParts, ", ")
}

func createMetricTable() {
	fmt.Println("Attempting to create metric table...")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("QDB_PG_USER"),
		os.Getenv("QDB_PG_PASSWORD"),
		os.Getenv("QUESTDB_HOST"),
		os.Getenv("QUESTDB_PORT"),
		"main")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to QuestDB:", err)
		return
	}
	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS driver_metrics (
		id TEXT,
		name TEXT,
		latitude DOUBLE PRECISION,
		longitude DOUBLE PRECISION,
		location_name TEXT,
		distance DOUBLE PRECISION,
		deliveries LONG,
		timestamp TIMESTAMP
	);`

	for retries := 0; retries < 5; retries++ {
		_, err = db.ExecContext(context.Background(), query)
		if err == nil {
			fmt.Println("Metric table created successfully")
			return
		}
		fmt.Printf("Attempt %d: Error creating table: %v\n", retries+1, err)
		time.Sleep(time.Second * 2)
	}
	fmt.Println("Failed to create metric table after multiple attempts")
}

func insertDriverMetrics() {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("QDB_PG_USER"),
		os.Getenv("QDB_PG_PASSWORD"),
		os.Getenv("QUESTDB_HOST"),
		os.Getenv("QUESTDB_PORT"),
		"main")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to QuestDB:", err)
		return
	}
	defer db.Close()

	for _, driver := range drivers {
		query := fmt.Sprintf(`INSERT INTO driver_metrics (id, name, latitude, longitude, location_name, distance, deliveries, timestamp) 
                  VALUES ('%s', '%s', %f, %f, '%s', %f, %d, '%s')`,
			driver.ID,
			driver.Name,
			driver.Location[0],
			driver.Location[1],
			driver.LocationName,
			driver.Distance,
			driver.Deliveries,
			time.Now().Format("2006-01-02 15:04:05"))
		_, err = db.ExecContext(context.Background(), query)
		if err != nil {
			fmt.Println("Error inserting data:", err)
		}
	}
}

func updateDriverLocations() {
	for {
		for i := range drivers {
			drivers[i].Location[0] += (rand.Float64() - 0.5) * 0.001
			drivers[i].Location[1] += (rand.Float64() - 0.5) * 0.001
			drivers[i].LocationName = getLocationName(drivers[i].Location[0], drivers[i].Location[1])
			drivers[i].Distance += rand.Float64() * 0.5
			drivers[i].Deliveries += rand.Intn(2)
		}
		time.Sleep(5 * time.Second)
	}
}

func getDriverMetrics(w http.ResponseWriter, r *http.Request) {
	for _, driver := range drivers {
		fmt.Fprintf(w, "Driver: %s, Location: %s, Distance: %.2f km, Deliveries: %d\n",
			driver.Name, driver.LocationName, driver.Distance, driver.Deliveries)
	}
}

func main() {
	createMetricTable()
	go updateDriverLocations()

	go func() {
		for {
			insertDriverMetrics()
			time.Sleep(10 * time.Second)
		}
	}()

	http.HandleFunc("/metrics", getDriverMetrics)

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
