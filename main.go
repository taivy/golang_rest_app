package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/validator.v2"
)

const (
	DB_PATH       = "./data.db"
	LOG_FILE_PATH = "log.log"
	LOGGING_LEVEL = log.InfoLevel
)

const tablesCreationQuery = `
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
email VARCHAR(100),
last_name VARCHAR(50),
first_name VARCHAR(50),
gender VARCHAR(1),
birth_date VARCHAR(25)
);

CREATE TABLE locations (
id INTEGER PRIMARY KEY AUTOINCREMENT,
place TEXT,
country VARCHAR(50),
city VARCHAR(50),
distance INT(32)
);

CREATE TABLE visits (
id INTEGER PRIMARY KEY AUTOINCREMENT,
location INT(32),
user INT(32),
visited_at VARCHAR(25),
mark INT(1),
FOREIGN KEY (location) REFERENCES locations(id),
FOREIGN KEY (user) REFERENCES users(id)
);
`

const (
	GET    = 0
	UPDATE = 1
)

type User struct {
	ID        int    `json:"id,omitempty"`
	Email     string `json:"email" validate:"nonzero"`
	FirstName string `json:"first_name" validate:"nonzero"`
	LastName  string `json:"last_name" validate:"nonzero"`
	Gender    string `json:"gender" validate:"nonzero"`
	BirthDate int    `json:"birth_date" validate:"nonzero"`
}

type Location struct {
	ID       int    `json:"id,omitempty"`
	Place    string `json:"place"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Distance int    `json:"distance"`
}

type Visit struct {
	ID        int    `json:"id,omitempty"`
	Location  int    `json:"location" validate:"nonzero"`
	User      int    `json:"user" validate:"nonzero"`
	VisitedAt string `json:"visited_at" validate:"nonzero"`
	Mark      int    `json:"mark" validate:"nonzero"`
}

func (User) TableName() string {
	return "users"
}

func (Visit) TableName() string {
	return "visits"
}

func (Location) TableName() string {
	return "locations"
}

var users []User

var mainDB *gorm.DB

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	if v[0] == "sql" {
		log.WithFields(log.Fields{"module": "gorm", "type": "sql"}).Print(v[3])
	}
	if v[0] == "log" {
		log.WithFields(log.Fields{"module": "gorm", "type": "log"}).Print(v[2])
	}
}

func InitDb() *gorm.DB {
	db, err := gorm.Open("sqlite3", DB_PATH)

	db.SetLogger(&GormLogger{})

	db.LogMode(true)

	if err != nil {
		panic(err)
	}

	return db
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func getEntities(w http.ResponseWriter, r *http.Request) {
	db := InitDb()
	defer db.Close()

	params := mux.Vars(r)

	var res interface{}
	entity, ok := params["entity"]
	if ok {
		entity = strings.ToLower(entity)
		switch entity {
		case "users":
			var foundEntities []User
			db.Find(&foundEntities)
			res = foundEntities
		case "visits":
			var foundEntities []Visit
			db.Find(&foundEntities)
			res = foundEntities
		case "locations":
			var foundEntities []Location
			db.Find(&foundEntities)
			res = foundEntities
		default:
			foundEntities := map[string]string{"Error": "Entity doesn't exist"}
			res = foundEntities
		}
	} else {
		res = map[string]string{"Error": "No entity specified"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func getOrUpdateEntity(entity string, id string, opType int, modelUpdates ...interface{}) (interface{}, int) {
	db := InitDb()
	defer db.Close()

	statusCode := 200

	var res interface{}
	switch entity {
	case "users":
		var foundEntity User
		db.Where("id = ?", id).First(&foundEntity)
		res = foundEntity
		if (foundEntity == User{}) {
			statusCode = 404
			break
		}
		if opType == UPDATE {
			db.Model(&foundEntity).Updates(modelUpdates[0])
		}
	case "visits":
		var foundEntity Visit
		db.Where("id = ?", id).First(&foundEntity)
		res = foundEntity
		if (foundEntity == Visit{}) {
			statusCode = 404
			break
		}
		if opType == UPDATE {
			db.Model(&foundEntity).Updates(modelUpdates[0])
		}
	case "locations":
		var foundEntity Location
		db.Where("id = ?", id).First(&foundEntity)
		res = foundEntity
		if (foundEntity == Location{}) {
			statusCode = 404
			break
		}
		if opType == UPDATE {
			db.Model(&foundEntity).Updates(modelUpdates[0])
		}
	default:
		res = map[string]string{"Error": "Entity type doesn't exist"}
		statusCode = 404
	}

	_, resIsSet := res.(map[string]string)
	// if code is 404 and res isn't set, set entity not found message
	if statusCode == 404 && !resIsSet {
		print()
		res = map[string]string{"Error": "Entity not found"}
	}
	return res, statusCode
}

func createEntity(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	db := InitDb()
	defer db.Close()

	w.Header().Set("Content-Type", "application/json; ")

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	check(err)

	body_ := bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var errUnmarshal error
	var errValidation error
	entity, ok := params["entity"]
	if ok {
		entity = strings.ToLower(entity)
		switch entity {
		case "users":
			var model User
			errUnmarshal = json.Unmarshal(body_, &model)
			errValidation = validator.Validate(model)
			if errUnmarshal == nil && errValidation == nil {
				db.Create(&model)
				json.NewEncoder(w).Encode(model)
			}
		case "visits":
			var model Visit
			errUnmarshal = json.Unmarshal(body_, &model)
			errValidation = validator.Validate(model)
			if errUnmarshal == nil && errValidation == nil {
				db.Create(&model)
				json.NewEncoder(w).Encode(model)
			}
		case "locations":
			var model Location
			//model = model.(Location)
			errUnmarshal = json.Unmarshal(body_, &model)
			errValidation = validator.Validate(model)
			if errUnmarshal == nil && errValidation == nil {
				db.Create(&model)
				json.NewEncoder(w).Encode(model)
			}
		default:
			res := map[string]string{"Error": "Entity doesn't exist"}
			json.NewEncoder(w).Encode(res)
			return
		}
	} else {
		res := map[string]string{"Error": "No entity specified"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if errUnmarshal != nil || errValidation != nil {
		w.WriteHeader(400)
		fmt.Println(errValidation)
		res := map[string]string{"Error": "Bad request body parameters"}
		json.NewEncoder(w).Encode(res)
	}
}

func deleteEntity(entity string, id string) (interface{}, int) {
	db := InitDb()
	defer db.Close()

	statusCode := 200

	var res interface{}
	res = map[string]interface{}{"Success": true}
	switch entity {
	case "users":
		var foundEntity User
		db.Where("id = ?", id).Delete(foundEntity)
	case "visits":
		var foundEntity Visit
		db.Where("id = ?", id).Delete(foundEntity)
	case "locations":
		var foundEntity Location
		db.Where("id = ?", id).Delete(foundEntity)
	default:
		res = map[string]string{"Error": "Entity doesn't exist"}
		statusCode = 404
	}
	return res, statusCode
}

func updateEntity(entity string, id string, rBody io.Reader) (interface{}, int) {
	db := InitDb()
	defer db.Close()

	body, err := ioutil.ReadAll(rBody)
	check(err)

	body_ := bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var errUnmarshal error
	var modelUpdated interface{}
	switch entity {
	case "users", "visits", "locations":
		errUnmarshal = json.Unmarshal(body_, &modelUpdated)
	default:
		res := map[string]string{"Error": "Entity doesn't exist"}
		return res, 404
	}

	nullFields := false
	for _, v := range modelUpdated.(map[string]interface{}) {
		if v == nil {
			nullFields = true
		}
	}

	var statusCode int
	var res interface{}
	if errUnmarshal != nil || nullFields {
		statusCode = 400
		res = map[string]string{"Error": "Bad request body parameters"}
	} else {
		res, statusCode = getOrUpdateEntity(entity, id, UPDATE, modelUpdated)
		if statusCode == 200 {
			res = map[string]interface{}{}
		}
	}

	return res, statusCode

}

func processEntity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	id, ok := params["id"]
	if !ok {
		res := map[string]string{"Error": "No ID specified"}
		json.NewEncoder(w).Encode(res)
		return
	}

	var res interface{}
	entity, ok := params["entity"]

	statusCode := 200
	if ok {
		entity = strings.ToLower(entity)
		switch r.Method {
		case http.MethodGet:
			res, statusCode = getOrUpdateEntity(entity, id, GET)
		case http.MethodPost:
			res, statusCode = updateEntity(entity, id, r.Body)
		case http.MethodDelete:
			res, statusCode = deleteEntity(entity, id)
		}
	} else {
		res = map[string]string{"Error": "No entity specified"}
		statusCode = 400
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(res)
}

func getUserVisits(w http.ResponseWriter, r *http.Request) {
	db := InitDb()
	defer db.Close()

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	id, ok := params["id"]
	if !ok {
		res := map[string]string{"Error": "No ID specified"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(res)
		return
	}

	qsError := true
	var err error
	qsParams := r.URL.Query()

	fromDate := qsParams.Get("fromDate")
	_, err = strconv.Atoi(fromDate)
	if err != nil {
		qsError = true
	}

	toDate := qsParams.Get("toDate")
	_, err = strconv.Atoi(toDate)
	if err != nil {
		qsError = true
	}

	country := qsParams.Get("country")

	toDistanceString := qsParams.Get("toDistance")
	toDistance, err := strconv.Atoi(toDistanceString)
	if err != nil {
		qsError = true
		toDistance = -1
	}

	if qsError {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"Error": "Bad query string parameters"})
		return
	}

	res, statusCode := getOrUpdateEntity("users", id, GET)
	if statusCode != 200 {
		json.NewEncoder(w).Encode(res)
		w.WriteHeader(statusCode)
		return
	}

	var visits []Visit
	db.Where("user = ?", id).Find(&visits)
	visitsFiltered := make([]Visit, 0)
	for _, v := range visits {
		model, statusCode := getOrUpdateEntity("locations", strconv.Itoa(v.Location), GET)
		var vLoc Location
		vLoc = model.(Location)
		if statusCode != 200 {
			continue
		}
		if (country == "" || vLoc.Country == country) && (fromDate == "" || v.VisitedAt > fromDate) && (toDate == "" || v.VisitedAt < toDate) && (toDistance == -1 || vLoc.Distance < toDistance) {
			visitsFiltered = append(visitsFiltered, v)
		}
	}
	json.NewEncoder(w).Encode(visitsFiltered)
}

func getUserAge(u User) int {
	ts := time.Unix(int64(u.BirthDate), 0)
	now := time.Now()
	y1, M1, _ := ts.Date()
	y2, M2, _ := now.Date()
	years := y2 - y1
	months := int(M2 - M1)
	if months < 0 {
		months += 12
		years--
	}
	return years
}

func filterVisitsGetMarks(id string, fromDate string, toDate string, fromAge int, toAge int, gender string) (int, int) {
	db := InitDb()
	defer db.Close()

	var visits []Visit
	db.Where("location = ?", id).Find(&visits)
	marksSum := 0
	marksCnt := 0
	for _, v := range visits {
		model, statusCode := getOrUpdateEntity("users", strconv.Itoa(v.User), GET)
		var vUser User
		vUser = model.(User)
		if statusCode != 200 {
			continue
		}
		var userAge int
		if fromAge != -1 {
			userAge = getUserAge(vUser)
		}
		if (gender == "" || vUser.Gender == gender) && (fromAge == -1 || userAge > fromAge) && (toAge == -1 || userAge < toAge) && (fromDate == "" || v.VisitedAt > fromDate) && (toDate == "" || v.VisitedAt < toDate) {
			marksSum += v.Mark
			marksCnt += 1
		}
	}
	return marksSum, marksCnt
}

func getLocationAvgMark(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	id, ok := params["id"]
	if !ok {
		res := map[string]string{"Error": "No ID specified"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(res)
		return
	}

	qsParams := r.URL.Query()

	fromDate := qsParams.Get("fromDate")
	toDate := qsParams.Get("toDate")
	gender := qsParams.Get("gender")

	fromAgeString := qsParams.Get("fromAge")
	fromAge, err := strconv.Atoi(fromAgeString)
	if err != nil {
		fromAge = -1
	}

	toAgeString := qsParams.Get("toAge")
	toAge, err := strconv.Atoi(toAgeString)
	if err != nil {
		toAge = -1
	}

	locFoundRes, statusCode := getOrUpdateEntity("locations", id, GET)
	if statusCode != 200 {
		if statusCode == 404 {
			// change "Entity not found" to "Location not found"
			locFoundRes = map[string]string{"Error": "Location not found"}
		}
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(locFoundRes)
		return
	}

	marksSum, marksCnt := filterVisitsGetMarks(id, fromDate, toDate, fromAge, toAge, gender)

	var avg float64
	if marksCnt == 0 {
		avg = 0
	} else {
		avg = float64(marksSum) / float64(marksCnt)
	}
	res := make(map[string]interface{})
	res["avg"] = math.Round(avg*10000) / 10000

	json.NewEncoder(w).Encode(res)
}

func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		// log request by who(IP address)
		requesterIP := r.RemoteAddr

		log.Printf(
			"%s\t\t%s\t\t%s\t\t%v",
			r.Method,
			r.RequestURI,
			requesterIP,
			time.Since(start),
		)
	})
}

func CreateDbIfNotExists() error {
	if _, err := os.Stat(DB_PATH); err == nil {
		return nil
	} else if os.IsNotExist(err) {
		// database not exists
		os.Create(DB_PATH)

		db, err := sql.Open("sqlite3", DB_PATH)
		if err != nil {
			return err
		}

		_, err = db.Exec(tablesCreationQuery)
		if err != nil {
			return err
		}

		db.Close()
		return nil
	} else {
		// database access error
		return err
	}
}

func ClearDB() {
	db := InitDb()
	defer db.Close()

	db.Delete(User{})
	db.Delete(Location{})
	db.Delete(Visit{})
}

func SetupHandlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/{entity}", getEntities).Methods("GET")
	r.HandleFunc("/{entity}/new", createEntity).Methods("POST")
	// get, update or delete
	r.HandleFunc("/{entity}/{id}", processEntity)
	r.HandleFunc("/users/{id}/visits", getUserVisits)
	r.HandleFunc("/locations/{id}/avg", getLocationAvgMark)
	return r
}

func main() {
	file, err := os.OpenFile(LOG_FILE_PATH, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	DBCreationErr := CreateDbIfNotExists()
	if DBCreationErr != nil {
		log.Fatal(DBCreationErr)
		panic(DBCreationErr)
	}

	ClearDB()

	log.SetOutput(file)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(LOGGING_LEVEL)

	r := SetupHandlers()

	log.Info("Server started")
	log.Fatal(http.ListenAndServe(":8000", RequestLogger(r)))

}
