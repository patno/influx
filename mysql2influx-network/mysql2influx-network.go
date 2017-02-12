package main

import (
	"fmt"
	"log"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/patno/influx/util"
)

const influxDatabase = "network"    // the database name
const influxMeasurement = "network" // the influx measurement

func main() {
	log.Println("network network ---------------------------------------")

	// Connect MySQL
	db := util.FactoryMySQL()
	defer db.Close()
	log.Println("Connected to MySQl")

	// Connect to InfluDB
	influx := util.FactoryInfluxDB()

	latestTimestamp := util.GetLatestTimestamp(influx, influxDatabase, influxMeasurement)
	log.Printf("Latest Influx DB timestamp:%v\n", latestTimestamp)

	// Read MySQL rows
	// mysqlDbQuery := "select timestamp, latency, download, upload from 1wire.network order by timestamp"

	mysqlDbQuery := fmt.Sprintf(
		"select timestamp, latency, download, upload from 1wire.network where timestamp > '%v' order by timestamp", latestTimestamp)

	log.Printf("MySQL Query:%v\n", mysqlDbQuery)
	rows, _, err := db.Query(mysqlDbQuery)
	util.CheckErr(err)
	numberOfRows := len(rows)
	log.Printf("Number of rows in MySQl:%v\n", numberOfRows)

	for _, row := range rows {
		timestampStr := row.Str(0)
		latency := row.Float(1)
		download := row.Float(2)
		upload := row.Float(3)
		timestamp := util.GetTimeFromString(timestampStr)
		log.Printf("ts=%v l=%v d=%v u=%v nts=%v\n", timestampStr, latency, download, upload, timestamp)
		//time.Sleep(10000 * time.Millisecond)

		// creating influx db data
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDatabase,
			Precision: "s",
		})
		util.CheckErr(err)

		tags := map[string]string{"method": "speedtest"}

		fields := map[string]interface{}{
			"latency":  latency,
			"download": download,
			"upload":   upload,
		}
		pt, err := client.NewPoint(influxMeasurement, tags, fields, timestamp)
		util.CheckErr(err)
		log.Println(pt)

		// writing data to influx
		bp.AddPoint(pt)
		influx.Write(bp)

	}

}
