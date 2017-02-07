package main

// import _ "github.com/go-sql-driver/mysql"
import "github.com/ziutek/mymysql/mysql"
import _ "github.com/ziutek/mymysql/native" // Native engine
import "fmt"
import "time"
import "github.com/pelletier/go-toml"
import "os"

const (
	username = "admin"
	password = "admin"
)

func main() {
	//user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true

	cfg, err := toml.LoadFile(os.Getenv("HOME") + "/config1wire.toml")
	checkErr(err)

	fmt.Println(cfg)

	db := mysql.New("tcp", "",
		cfg.Get("mysql.host").(string),
		cfg.Get("mysql.user").(string),
		cfg.Get("mysql.password").(string),
		cfg.Get("mysql.database").(string))

	err = db.Connect()
	if err != nil {
		panic(err)
	}

	//select unix_timestamp(timestamp) * 1000000000 as datetime, latency, download, upload from 1wire.network order by datetime

	defer db.Close()
	fmt.Println("Connected:")
	fmt.Println(db)

	rows, _, err := db.Query("select timestamp, latency, download, upload from 1wire.network order by timestamp")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	layout := "2006-01-02 15:04:05"
	numRows := len(rows)
	for _, row := range rows {
		timestampStr := row.Str(0)
		latency := row.Float(1)
		download := row.Float(2)
		upload := row.Float(3)
		timestamp, err := time.Parse(layout, timestampStr)
		checkErr(err)
		fmt.Printf("%v %v %v %v %v\n", timestampStr, latency, download, upload, timestamp)
	}
	fmt.Printf("Number of rows: %v\n", numRows)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
