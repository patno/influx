package main

// import _ "github.com/go-sql-driver/mysql"

import ( // Native engine
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pelletier/go-toml"
	_ "github.com/ziutek/mymysql/native"

	util "github.com/patno/influx/util"
)

func main() {
	//user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true

	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	util.CheckErr(err)
	fmt.Println(cfg)

	db := util.FactoryMySQL()
	//select unix_timestamp(timestamp) * 1000000000 as datetime, latency, download, upload from 1wire.network order by datetime

	defer db.Close()
	log.Println("Connected:")
	log.Println(db)

	rows, _, err := db.Query("select timestamp, latency, download, upload from 1wire.network order by timestamp")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	numRows := len(rows)
	for _, row := range rows {
		timestampStr := row.Str(0)
		latency := row.Float(1)
		download := row.Float(2)
		upload := row.Float(3)
		timestamp, err := time.Parse(util.LayoutMYSQLDate, timestampStr)
		util.CheckErr(err)
		fmt.Printf("%v %v %v %v %v\n", timestampStr, latency, download, upload, timestamp)
	}

	log.Printf("Number of rows: %v\n", numRows)
}
