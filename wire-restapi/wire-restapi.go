package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/patno/influx/util"
)

type QueryResponseAPI struct {
	Target     string       `json:"target,omitempty"`
	Datapoints [][2]float64 `json:"datapoints",omitempty`
}

// from Macbook Air
type Target struct {
	Target string `json:"target"`
	RefID  int    `json:"refId"`
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
	Interval   string   `json:"interval"`
	IntervalMs int      `json:"intervalMs"`
	Targets    []Target `json:"targets"`
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

	queryResponse = append(queryResponse, "download")
	queryResponse = append(queryResponse, "upload")
	queryResponse = append(queryResponse, "latency")
	for key := range util.NameToDeviceID {
		queryResponse = append(queryResponse, key)
	}
	json.NewEncoder(w).Encode(queryResponse)
}

func postAnnotationsAPI(w http.ResponseWriter, req *http.Request) {
	log.Println("wire rest API Annotations")
	setCORSHeaders(w)
	log.Printf("Annotations:\n%+v\n", req)
	//json.NewEncoder(w).Encode(queryResponse)
}

func getQueryAPI(w http.ResponseWriter, request *http.Request) {
	var qr [1]QueryResponseAPI
	var req QueryRequestAPI
	log.Println("wire rest API Query")

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

	targetIndex := 0
	// Loop the targets that are queried for from grafana.
	for _, targetQuery := range req.Targets {
		ID := util.NameToDeviceID[targetQuery.Target]
		log.Printf("ID to query:%v\n", ID)

		var mysqlDbQuery string

		if strings.HasPrefix(ID, "10") {
			mysqlDbQuery = fmt.Sprintf(
				"select timestamp, value from 1wire.energi where timestamp > '%v' and timestamp < '%v' and deviceid = '%v' order by timestamp",
				startTime.Format(util.LayoutMYSQLDate),
				endTime.Format(util.LayoutMYSQLDate), ID)
		} else {
			mysqlDbQuery = fmt.Sprintf(
				"select timestamp, %v from 1wire.network where timestamp > '%v' and timestamp < '%v' and order by timestamp",
				ID,
				startTime.Format(util.LayoutMYSQLDate),
				endTime.Format(util.LayoutMYSQLDate))
		}

		log.Printf("MySQL Query:%v\n", mysqlDbQuery)

		rows, _, err := db.Query(mysqlDbQuery)
		util.CheckErr(err)
		log.Printf("Number of rows in MySQl:%v\n", len(rows))

		qr[targetIndex].Target = targetQuery.Target
		qr[targetIndex].Datapoints = make([][2]float64, len(rows))
		i := 0

		for _, row := range rows {
			timestampStr := row.Str(0)
			value := row.Float(1)

			//log.Printf("t=%v, v=%v, i=%v", timestampStr, value, i)

			if timestampStr == "0000-00-00 00:00:00" {
				continue
			}

			timestamp := util.GetTimeFromString(timestampStr, util.LayoutMYSQLDate)
			qr[i].Datapoints[i][1] = float64(timestamp.Unix() * 1000)
			qr[i].Datapoints[i][0] = value
			i++
		}
		targetIndex++
	}

	log.Printf("JSON Response:\n%+v\n", qr)
	json.NewEncoder(w).Encode(qr)
}

func main() {
	router := mux.NewRouter()
	port := ":8346"
	log.Printf("1wire and network rest API %v-------------------\n", port)
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
