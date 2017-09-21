package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/nakagami/firebirdsql"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// StorageItem represents row from database
type StorageItem struct {
	Group    string `json:"group"`
	Name     string `json:"name"`
	Article  string `json:"article"`
	Price    string `json:"price"`
	Quantity int    `json:"quantity"`
}

// Configuration represents json configuration file
type Configuration struct {
	Connection string
	DbPath     string
	RequestURL string
	Login      string
	Password   string
}

func main() {
	// Init
	config := readConfig()
	storage := getStorage(config.Connection, config.DbPath)
	payload, _ := json.Marshal(storage)
	method := "http"

	// Get argument, if presented
	if len(os.Args) > 1 {
		method = os.Args[1]
	}

	// Switch methods, file or http supported
	if method == "file" {
		writeToFile(payload)
	} else if method == "http" {
		sendRequest(config, payload, config.RequestURL)
	} else {
		log.Fatal("Unknown argument: ", method)
	}
}

// Read configuration from json
func readConfig() Configuration {
	fmt.Println("Reading configuration file")

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filepath := []string{dir, "config.json"}

	file, _ := os.Open(strings.Join(filepath, "\\"))
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	if err != nil {
		log.Fatal(err)
	}

	return configuration
}

// Get storage inventory from shop database
func getStorage(connection string, dbPath string) []*StorageItem {
	fmt.Println("Connecting to database using", dbPath)

	conn := []string{connection, "/", dbPath}
	db, err := sql.Open("firebirdsql", strings.Join(conn, ""))
	if err != nil {
		log.Fatal(err)
	}

	// Main query, get products
	rows, err := db.Query("SELECT GR.GROUP_NAME, G.GOODS_NAME, G.GOODS_ARTICLE, cast(G.GOODS_COST as varchar(100)), G.GOODS_COUNT FROM GETGOODSINSTORAGENESTED(1,1,1,1,1,2,CURRENT_DATE,0) G LEFT JOIN GOODS_GROUP GR ON GR.GROUP_ID = G.GROUP_ID WHERE G.GOODS_ARTICLE != '' AND G.GOODS_COST > 0 AND G.GOODS_COUNT > 0")

	if err != nil {
		fmt.Println("Connection error!")
		log.Fatal(err)
	}

	defer rows.Close()

	storageItems := make([]*StorageItem, 0)

	// Create StorageItem for each row
	for rows.Next() {
		bk := new(StorageItem)
		err := rows.Scan(&bk.Group, &bk.Name, &bk.Article, &bk.Price, &bk.Quantity)
		if err != nil {
			log.Fatal(err)
		}

		bk.Group = toUtf(bk.Group)
		bk.Name = toUtf(bk.Name)
		bk.Article = toUtf(bk.Article)

		storageItems = append(storageItems, bk)
	}

	fmt.Println("Total products found:", len(storageItems))

	return storageItems
}

// Convert ugly windows 1251 to utf
func toUtf(input string) string {
	sr := strings.NewReader(input)
	tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil {
		log.Fatal(err)
	}

	s := string(buf)

	return s
}

// Write bytes to json file
func writeToFile(payload []byte) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filepath := []string{dir, "export.json"}

	destination := strings.Join(filepath, "\\")

	fmt.Println("Writing data to file", destination)
	err := ioutil.WriteFile(destination, payload, 0644)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done")
}

// Send HTTP POST request to import endpoint
func sendRequest(config Configuration, payload []byte, url string) {
	fmt.Println("Sending request to remote...")
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	request.SetBasicAuth(config.Login, config.Password)
	resp, err := client.Do(request)

	if err != nil {
		fmt.Println("Request connection error!")
		log.Fatal(err)
	}

	fmt.Println("Request sent. Status:", resp.Status)
}
