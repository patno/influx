package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

type QueryResponseAPI struct {
	Target     string     `json:"target,omitempty"`
	Datapoints [][2]int64 `json:"datapoints",omitempty`
}

type QueryRequestAPI struct {
	PanelID int `json:"panelId"`
	Range   struct {
		From string `json:"from"`
		To   string `json:"to"`
		Raw  struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"raw"`
	} `json:"range"`
	RangeRaw struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"rangeRaw"`
	Interval   string `json:"interval"`
	IntervalMs int    `json:"intervalMs"`
}

func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(people)
}

func CreatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

func DeletePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(people)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "keep-alive")
}

func getTestConnectionAPI(w http.ResponseWriter, req *http.Request) {
	log.Println("energy rest API Test Connection")
	setCORSHeaders(w)
	//w.WriteHeader(http.StatusNotModified)

	key := "black"
	e := `"` + key + `"`
	w.Header().Set("Etag", e)
	w.Header().Set("Cache-Control", "max-age=2592000") // 30 days

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, e) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

}

func getSearchAPI(w http.ResponseWriter, req *http.Request) {
	var queryResponse []string
	log.Printf("energy rest API Search:%+v\n", req)
	setCORSHeaders(w)
	queryResponse = append(queryResponse, "energy")
	json.NewEncoder(w).Encode(queryResponse)
}

func postAnnotationsAPI(w http.ResponseWriter, req *http.Request) {
	log.Println("energy rest API Annotations")
	setCORSHeaders(w)
	log.Printf("Annotations:\n%+v\n", req)
	//json.NewEncoder(w).Encode(queryResponse)
}

func getQueryAPI(w http.ResponseWriter, request *http.Request) {
	var qr [1]QueryResponseAPI
	var req QueryRequestAPI
	log.Println("energy rest API Query")
	setCORSHeaders(w)

	_ = json.NewDecoder(request.Body).Decode(&req)
	log.Printf("HTTP Request:\n%+v\n", request)
	log.Printf("JSON Request:\n%+v\n", req)
	qr[0].Target = "energy"
	qr[0].Datapoints = make([][2]int64, 1)
	qr[0].Datapoints[0][0] = 1
	qr[0].Datapoints[0][1] = 2
	log.Printf("JSON Response:\n%+v\n", qr)
	json.NewEncoder(w).Encode(qr)
}

func main() {
	router := mux.NewRouter()
	people = append(people, Person{ID: "1", Firstname: "Nic", Lastname: "Raboy", Address: &Address{City: "Dublin", State: "CA"}})
	people = append(people, Person{ID: "2", Firstname: "Maria", Lastname: "Raboy"})
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people/{id}", DeletePersonEndpoint).Methods("DELETE")
	log.Println("energy rest API -------------------")
	router.HandleFunc("/", getTestConnectionAPI)
	router.HandleFunc("/", getTestConnectionAPI)
	router.HandleFunc("/search", getSearchAPI)
	router.HandleFunc("/query", getQueryAPI)
	router.HandleFunc("/annotations", postAnnotationsAPI)

	err := http.ListenAndServe(":8345", router)
	if err != nil {
		log.Fatal(err)
	}
}
