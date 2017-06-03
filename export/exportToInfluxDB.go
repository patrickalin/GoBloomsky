package export

import (
	"fmt"
	"time"

	clientinfluxdb "github.com/influxdata/influxdb/client/v2"
	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
	config "github.com/patrickalin/bloomsky-client-go/config"
)

type influxDBStruct struct {
	Columns []string        `json:"columns"`
	Serie   string          `json:"name"`
	Points  [][]interface{} `json:"points"`
}

var clientInflux clientinfluxdb.Client

func sendbloomskyToInfluxDB(onebloomsky bloomskyStructure.BloomskyStructure, oneConfig config.Configuration) {

	fmt.Printf("\n%s :> Send bloomsky Data to InfluxDB\n", time.Now().Format(time.RFC850))

	// Create a point and add to batch
	tags := map[string]string{"bloomsky": "living"}
	fields := map[string]interface{}{
		"NumOfFollowers": onebloomsky.GetNumOfFollowers(),
	}

	// Create a new point batch
	bp, err := clientinfluxdb.NewBatchPoints(clientinfluxdb.BatchPointsConfig{
		Database:  oneConfig.InfluxDBDatabase,
		Precision: "s",
	})

	if err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error sent Data to Influx DB : %v", err))
	}

	pt, err := clientinfluxdb.NewPoint("bloomskyData", tags, fields, time.Now())
	bp.AddPoint(pt)

	// Write the batch
	err = clientInflux.Write(bp)

	if err != nil {
		err2 := createDB(oneConfig)
		if err2 != nil {
			mylog.Error.Fatal(fmt.Errorf("Check if InfluxData is running or if the database bloomsky exists : %v", err))
		}
	}
}

func createDB(oneConfig config.Configuration) error {
	fmt.Println("Create Database bloomsky in InfluxData")

	query := fmt.Sprint("CREATE DATABASE ", oneConfig.InfluxDBDatabase)
	q := clientinfluxdb.NewQuery(query, "", "")

	fmt.Println("Query: ", query)

	_, err := clientInflux.Query(q)
	if err != nil {
		return fmt.Errorf("Error with : Create database bloomsky, check if InfluxDB is running : %v", err)
	}
	fmt.Println("Database bloomsky created in InfluxDB")
	return nil
}

func makeClient(oneConfig config.Configuration) (client clientinfluxdb.Client, err error) {
	client, err = clientinfluxdb.NewHTTPClient(
		clientinfluxdb.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", oneConfig.InfluxDBServer, oneConfig.InfluxDBServerPort),
			Username: oneConfig.InfluxDBUsername,
			Password: oneConfig.InfluxDBPassword,
		})

	if err != nil || client == nil {
		return nil, fmt.Errorf("Error creating database bloomsky, check if InfluxDB is running : %v", err)
	}
	return client, nil
}

// InitInfluxDB initiate the client influxDB
// Arguments bloomsky informations, configuration from config file
// Wait events to send to influxDB
func InitInfluxDB(messagesbloomsky chan bloomskyStructure.BloomskyStructure, oneConfig config.Configuration) {

	clientInflux, _ = makeClient(oneConfig)

	go func() {
		mylog.Trace.Println("Receive messagesbloomsky to export InfluxDB")
		for {
			msg := <-messagesbloomsky
			sendbloomskyToInfluxDB(msg, oneConfig)
		}
	}()

}
