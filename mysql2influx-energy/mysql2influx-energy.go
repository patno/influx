package main

import (
	"fmt"
	"log"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/patno/influx/util"
)

const influxDatabase = "testdb"    // the database name
const influxMeasurement = "energy" // the influx measurement

func main() {
	log.Println("1wire energy -------------------------------------")

	// Connect MySQL
	db := util.FactoryMySQL()
	defer db.Close()
	log.Println("Connected to MySQl")

	// Connect to InfluDB
	influx := util.FactoryInfluxDB(influxDatabase)

	latestTimestamp := util.GetLatestTimestamp(influx, influxDatabase, influxMeasurement)
	log.Printf("Latest Influx DB timestamp:%v\n", latestTimestamp)

	// Read MySQL rows
	mysqlDbQuery := fmt.Sprintf(
		"select timestamp, deviceid, value from 1wire.network where timestamp > '%v' order by timestamp", latestTimestamp)

	log.Printf("MySQL Query:%v\n", mysqlDbQuery)
	rows, _, err := db.Query(mysqlDbQuery)
	util.CheckErr(err)
	numberOfRows := len(rows)
	log.Printf("Number of rows in MySQl:%v\n", numberOfRows)

	for _, row := range rows {
		timestampStr := row.Str(0)
		deviceID := row.Str(1)
		value := row.Float(2)

		if timestampStr == "0000-00-00 00:00:00" {
			continue
		}

		timestamp := util.GetTimeFromString(timestampStr)
		deviceName := "energy"

		log.Printf("ts=%v id=%v n=%v v=%v nts=%v\n", timestampStr, deviceID, deviceName, value, timestamp)
		time.Sleep(10000 * time.Millisecond)

		// creating influx db data
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDatabase,
			Precision: "s",
		})
		util.CheckErr(err)
		tags := map[string]string{"id": deviceID, "name": deviceName}

		fields := map[string]interface{}{
			"value": value,
		}
		pt, err := client.NewPoint(influxMeasurement, tags, fields, timestamp)
		util.CheckErr(err)
		log.Println(pt)

		// writing data to influx
		bp.AddPoint(pt)
		influx.Write(bp)

	}
}
