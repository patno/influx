package util

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/ziutek/mymysql/mysql"

	"os"

	"fmt"

	"github.com/pelletier/go-toml"
	_ "github.com/ziutek/mymysql/native" // will not work otherwise
)

// Native engine

// NameToDeviceID Friendy name to ID mapping
var NameToDeviceID = map[string]string{
	"Utetemp":      "10.1C0D2A020800",
	"RadiatorVarm": "10.14012A020800",
	"Källare":      "10.8F4F29020800",
	"RadiatorKall": "10.6E4F29020800",
	"VattenVarm":   "10.1C012A020800",
	"VattenKall":   "10.771B2A020800",
	"Mellanvåning": "10.08FE29020800",
	"HåletKall":    "10.BD0A2A020800",
	"HåletVarm":    "10.79172A020800",
}

// LayoutMYSQLDate layout of MySQL date
const LayoutMYSQLDate = "2006-01-02 15:04:05"

// LayoutSimpleJSONQueryDate layout of dates when simple-json queries from grafana
// Example string 2017-03-07T23:38:57.112Z
const LayoutSimpleJSONQueryDate = "2006-01-02T15:04:05.000Z"

// GetTimeFromString parses date time string. Panics if fails.
func GetTimeFromString(timestampStr string, layout string) time.Time {
	t, err := time.Parse(layout, timestampStr)
	CheckErr(err)
	return t
}

// GetLatestTimestamp reads out the timestamp for the last row. Panics if fails.
func GetLatestTimestamp(c client.Client, db string, measurement string) string {
	query := fmt.Sprintf("SELECT * FROM %v GROUP BY * ORDER BY DESC LIMIT 1", measurement)
	res, err := QueryDB(c, query, db)
	CheckErr(err)
	if len(res[0].Series) == 0 {
		return "2000-01-01 00:00:00"
	}

	var myData [][]interface{} = make([][]interface{}, len(res[0].Series[0].Values))
	for i, d := range res[0].Series[0].Values {
		myData[i] = d
	}

	lastTime := myData[0][0].(string)
	return lastTime
}

// QueryDB do a generic query towards influxdb
func QueryDB(clnt client.Client, cmd string, MyDB string) (res []client.Result, err error) {
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

// FactoryMySQL creates a database connection to MySQL. Panics on failure
func FactoryMySQL() mysql.Conn {

	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	CheckErr(err)

	//log.Println(cfg)

	db := mysql.New("tcp", "",
		cfg.Get("mysql.host").(string),
		cfg.Get("mysql.user").(string),
		cfg.Get("mysql.password").(string),
		cfg.Get("mysql.database").(string))

	err = db.Connect()

	if err != nil {
		panic(err)
	}
	return db

}

// FactoryInfluxDB creates a client to InfluxDB. Panics is fails.
func FactoryInfluxDB(database string) client.Client {
	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	CheckErr(err)

	// preparing influx db
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     cfg.Get("influxdb.url").(string),
		Username: cfg.Get("influxdb.user").(string),
		Password: cfg.Get("influxdb.password").(string),
	})
	CheckErr(err)
	if database != "" {
		q := client.NewQuery("CREATE DATABASE "+database, "", "")
		if response, err := c.Query(q); err == nil && response.Error() == nil {
			fmt.Println(response.Results)
		}
	}
	return c
}

// CheckErr checks for errors and panics if any
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
