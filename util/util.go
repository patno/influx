package util

import (
	"log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/ziutek/mymysql/mysql"

	"os"

	"github.com/pelletier/go-toml"
	_ "github.com/ziutek/mymysql/native" // will not work otherwise
)

// Native engine

// LayoutMYSQLDate layout of MySQL date
const LayoutMYSQLDate = "2006-01-02 15:04:05"

// GetTimeFromString parses date time string. Panics if fails.
func GetTimeFromString(timestampStr string) time.Time {
	t, err := time.Parse(LayoutMYSQLDate, timestampStr)
	CheckErr(err)
	return t
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

	log.Println(cfg)

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
func FactoryInfluxDB() client.Client {
	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	CheckErr(err)

	// preparing influx db
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     cfg.Get("influxdb.url").(string),
		Username: cfg.Get("influxdb.user").(string),
		Password: cfg.Get("influxdb.password").(string),
	})
	CheckErr(err)
	return c
}

// CheckErr checks for errors and panics if any
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
