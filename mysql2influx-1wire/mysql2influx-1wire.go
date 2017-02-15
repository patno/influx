package main

import (
	"fmt"
	"log"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/patno/influx/util"
)

const influxDatabase = "testdb"         // the database name
const influxMeasurement = "temperature" // the influx measurement

func main() {
	log.Println("1wire temperature -------------------------------------")

	// Connect MySQL
	db := util.FactoryMySQL()
	defer db.Close()
	log.Println("Connected to MySQl")

	// Connect to InfluDB
	influx := util.FactoryInfluxDB()

	latestTimestamp := util.GetLatestTimestamp(influx, influxDatabase, influxMeasurement)
	log.Printf("Latest Influx DB timestamp:%v\n", latestTimestamp)

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
		deviceName := deviceID2deviceName(deviceID)

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

func deviceID2deviceName(id string) string {
	switch id {
	case "10.1C0D2A020800":
		return "Utetemp"
	case "10.14012A020800":
		return "RadiatorVarm"
	case "10.8F4F29020800":
		return "Källare"
	case "10.6E4F29020800":
		return "RadiatorKall"
	case "10.1C012A020800":
		return "VattenVarm"
	case "10.771B2A020800":
		return "VattenKall"
	case "10.08FE29020800":
		return "Mellanvåning"
	case "10.BD0A2A020800":
		return "HåletKall"
	case "10.79172A020800":
		return "HåletVarm"
	}
	panic(fmt.Sprintf("No matching device name for id:%v\n", id))
}
