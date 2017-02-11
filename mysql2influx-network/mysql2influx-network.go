package main

import (
	"fmt"
	"log"

	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/patno/influx/util"
)

func main() {
	log.Println("-------------------------------------------------------")

	// Connect MySQL
	db := util.FactoryMySQL()
	defer db.Close()
	log.Println("Connected to MySQl")

	// Read MySQL rows
	rows, _, err := db.Query("select timestamp, latency, download, upload from 1wire.network order by timestamp")
	util.CheckErr(err)
	numberOfRows := len(rows)
	log.Printf("Number of rows in MySQl:%v\n", numberOfRows)

	// Connect to InfluDB
	influx := util.FactoryInfluxDB()

	for _, row := range rows {
		timestampStr := row.Str(0)
		latency := row.Float(1)
		download := row.Float(2)
		upload := row.Float(3)
		timestamp := util.GetTimeFromString(timestampStr)
		fmt.Printf("%v %v %v %v %v\n", timestampStr, latency, download, upload, timestamp)
		time.Sleep(2000 * time.Millisecond)

		// creating influx db data
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "testdb",
			Precision: "s",
		})
		util.CheckErr(err)

		tags := map[string]string{"method": "speedtest"}

		fields := map[string]interface{}{
			"latency":  latency,
			"download": download,
			"upload":   upload,
		}
		pt, err := client.NewPoint("network", tags, fields, timestamp)
		util.CheckErr(err)
		log.Println(pt)

		// writing data to influx
		bp.AddPoint(pt)
		influx.Write(bp)

	}

}
