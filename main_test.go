package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var r *mux.Router
var db *gorm.DB

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func TestMain(m *testing.M) {
	DBCreationErr := CreateDbIfNotExists()
	if DBCreationErr != nil {
		panic(DBCreationErr)
	}

	r = SetupHandlers()
	db = InitDb()
	db.LogMode(false)
	ClearDB()

	os.Exit(m.Run())
	db.Close()
}

func TestGetNonExistentEntity(t *testing.T) {
	ClearDB()
	req, _ := http.NewRequest("GET", "/badentity/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity type doesn't exist" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Entity not found'. Got '%s'", m["error"])
	}
}

func TestGetNonExistentUser(t *testing.T) {
	ClearDB()
	req, _ := http.NewRequest("GET", "/users/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Entity not found'. Got '%s'", m["error"])
	}
}

func TestGetNonExistentVisit(t *testing.T) {
	ClearDB()
	req, _ := http.NewRequest("GET", "/visits/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Entity not found'. Got '%s'", m["error"])
	}
}

func TestGetNonExistentLocation(t *testing.T) {
	ClearDB()
	req, _ := http.NewRequest("GET", "/locations/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Entity not found'. Got '%s'", m["error"])
	}
}

func TestUpdateNonExistentUser(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "first_name": "Jack"
    }
    `)

	req, _ := http.NewRequest("POST", "/users/1", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
	}
}

func TestCreateUser(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "id": 1,
        "email": "johsmith@mail.com",
        "first_name": "John",
        "last_name": "Smith",
        "gender": "m",
        "birth_date": 1290129012
    }
    `)
	req, _ := http.NewRequest("POST", "/users/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateUserWithNullField(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "id": 1,
        "email": null,
        "first_name": "John",
        "last_name": "Smith",
        "gender": "m",
        "birth_date": 1290129012
    }
    `)
	req, _ := http.NewRequest("POST", "/users/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Bad request body parameters" {
		t.Errorf("Expected the 'Error' key of the response to be set to 'Bad request body parameters'. Got '%s'", m["Error"])
	}
}

func TestCreateUserWithIncompleteFields(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "id": 1,
        "first_name": "John",
        "last_name": "Smith",
        "birth_date": 1290129012
    }
    `)
	req, _ := http.NewRequest("POST", "/users/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Bad request body parameters" {
		t.Errorf("Expected the 'Error' key of the response to be set to 'Bad request body parameters'. Got '%s'", m["Error"])
	}
}

func TestGetExistentUser(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "id": 1,
        "email": "johsmith@mail.com",
        "first_name": "John",
        "last_name": "Smith",
        "gender": "m",
        "birth_date": 1290129012
    }
    `)

	req, _ := http.NewRequest("POST", "/users/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/users/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["first_name"] != "John" {
		t.Errorf("Expected the 'first_name' key of the response to be set to 'John'. Got '%s'", m["Error"])
	}

	if m["last_name"] != "Smith" {
		t.Errorf("Expected the 'last_name' key of the response to be set to 'Smith'. Got '%s'", m["Error"])
	}
}

func TestUpdateUser(t *testing.T) {
	payload := []byte(`
    {
        "first_name": "Jack"
    }
    `)

	req, _ := http.NewRequest("POST", "/users/1", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/users/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["first_name"] != "Jack" {
		t.Errorf("Expected the 'first_name' key of the response to be set to 'Jack'. Got '%s'", m["Error"])
	}
}

func TestCreateLocation(t *testing.T) {
	ClearDB()
	payload := []byte(`
	{
	    "id": 1,
	    "place": "Red Square",
	    "country": "Russia",
	    "city": "Moscow",
	    "distance": 29192190
	}
    `)
	req, _ := http.NewRequest("POST", "/locations/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateVisit(t *testing.T) {
	ClearDB()
	payload := []byte(`
	{
	    "id": 1,
	    "location": 1,
	    "user": 1,
	    "visited_at": "365299700",
	    "mark": 5
	}
    `)
	req, _ := http.NewRequest("POST", "/visits/new", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateUserWithNullFields(t *testing.T) {
	ClearDB()
	payload := []byte(`
    {
        "first_name": null
    }
    `)

	req, _ := http.NewRequest("POST", "/users/1", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Bad request body parameters" {
		t.Errorf("Expected the 'Error' key of the response to be set to 'Bad request body parameters'. Got '%s'", m["Error"])
	}
}

func TestGetUserVisitsWithWrongArgsInQueryString(t *testing.T) {
	req, _ := http.NewRequest("GET", `/users/1/visits?fromDate=abracadbra`, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Bad query string parameters" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Bad query string parameters'. Got '%s'", m["error"])
	}
}

func TestGetUserVisitsWithNullArgsInQueryString(t *testing.T) {
	req, _ := http.NewRequest("GET", `/users/1/visits?fromDate=`, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Bad query string parameters" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Bad query string parameters'. Got '%s'", m["error"])
	}
}
func TestGetUserVisitsNonExistingUser(t *testing.T) {
	req, _ := http.NewRequest("GET", `/users/99999/visits`, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["Error"] != "Entity nof found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Entity nof found'. Got '%s'", m["error"])
	}
}
