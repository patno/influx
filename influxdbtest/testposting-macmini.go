package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	util "github.com/patno/influx/util"
	"github.com/pelletier/go-toml"
)

func main() {
	now := time.Now()
	// load configuration
	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	if err != nil {
		log.Panic(err)
	}

	// preparing influx db
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     cfg.Get("influxdb.url").(string),
		Username: cfg.Get("influxdb.user").(string),
		Password: cfg.Get("influxdb.password").(string),
	})

	if err != nil {
		log.Fatal("Error:", err)
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "testdb",
		Precision: "s",
	})

	if err != nil {
		log.Fatal("Error:", err)
	}

	// creating influx db data
	tags := map[string]string{"cpu:": "cpu-total"}

	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}
	pt, err := client.NewPoint("cpu_usage", tags, fields, now)

	if err != nil {
		log.Fatalln("Error: ", err)
	}
	log.Println(pt)

	// writing data to influx
	bp.AddPoint(pt)
	c.Write(bp)

	// Query influx about latest data point
	log.Println("Quering the Database")
	res, err := util.QueryDB(c, "SELECT * FROM cpu_usage GROUP BY * ORDER BY DESC LIMIT 1", "testdb")

	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println(res)
	log.Println("---------------")
	log.Println(res[0].Series[0].Values[0])

	var myData [][]interface{} = make([][]interface{}, len(res[0].Series[0].Values))
	for i, d := range res[0].Series[0].Values {
		myData[i] = d
	}

	fmt.Println("1:", myData[0]) //first element in slice
	fmt.Println("2:", myData[0][0])
	fmt.Println("3:", myData[0][1])

	lastTime := myData[0][0].(string)

	fmt.Println("lastTime:" + lastTime)
}

/*
func queryDB(clnt client.Client, cmd string, MyDB string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: MyDB,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
*/
