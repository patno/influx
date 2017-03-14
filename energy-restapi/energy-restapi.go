package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"math"

	"github.com/gorilla/mux"
	"github.com/patno/influx/util"
)

type QueryResponseAPI struct {
	Target     string       `json:"target,omitempty"`
	Datapoints [][2]float64 `json:"datapoints",omitempty`
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
	//log.Printf("HTTP Request:\n%+v\n", request)
	log.Printf("JSON Request:\n%+v\n", req)

	if req.Range.From == "" {
		return
	}
	startTime := util.GetTimeFromString(req.Range.From, util.LayoutSimpleJSONQueryDate)
	endTime := util.GetTimeFromString(req.Range.To, util.LayoutSimpleJSONQueryDate)

	db := util.FactoryMySQL()
	defer db.Close()

	mysqlDbQuery := fmt.Sprintf(
		"select timestamp, deviceid, value from 1wire.energi where timestamp > '%v' and timestamp < '%v' order by timestamp",
		startTime.Format(util.LayoutMYSQLDate),
		endTime.Format(util.LayoutMYSQLDate))

	log.Printf("MySQL Query:%v\n", mysqlDbQuery)

	rows, _, err := db.Query(mysqlDbQuery)
	util.CheckErr(err)
	log.Printf("Number of rows in MySQl:%v\n", len(rows))

	qr[0].Target = "energy"
	qr[0].Datapoints = make([][2]float64, len(rows))
	i := 0
	oldy := 0.0
	//	oldx := int64(0)
	for _, row := range rows {
		timestampStr := row.Str(0)
		value := row.Float(2)

		//log.Printf("t=%v, v=%v, i=%v", timestampStr, value, i)

		if timestampStr == "0000-00-00 00:00:00" {
			continue
		}
		i++
		if i == 1 {
			oldy = value
			continue
		}

		timestamp := util.GetTimeFromString(timestampStr, util.LayoutMYSQLDate)
		diffy := value - oldy
		//diffx := timestamp.Unix()*1000 - oldx

		if math.Abs(diffy) > 1000 {
			continue
		}

		qr[0].Datapoints[i-1][1] = float64(timestamp.Unix() * 1000)
		qr[0].Datapoints[i-1][0] = diffy
		oldy = value
	}
	log.Printf("JSON Response:\n%+v\n", qr)
	json.NewEncoder(w).Encode(qr)
}

func main() {
	router := mux.NewRouter()
	port := ":8345"
	log.Printf("energy rest API %v-------------------\n", port)
	router.HandleFunc("/", getTestConnectionAPI)
	router.HandleFunc("/", getTestConnectionAPI)
	router.HandleFunc("/search", getSearchAPI)
	router.HandleFunc("/query", getQueryAPI)
	router.HandleFunc("/annotations", postAnnotationsAPI)

	err := http.ListenAndServe(port, router)
	if err != nil {
		log.Fatal(err)
	}
}
