package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
)

const (
	nameFileJSON = "db.json"
	nameFileCSV  = "db.csv"
)

type product struct {
	Product string `json:"product"`
	Price   int    `json:"price"`
	Rating  int    `json:"rating"`
}

func main() {
	resultsProcessing := make(chan product, 4)
	stopChan := make(chan struct{}, 1)

	go fileProcessingJSON(resultsProcessing, stopChan)
	go fileProcessingCSV(resultsProcessing, stopChan)

	var resProcessing []product
	for i := 0; i < 4; i++ {
		select {
		case resProd := <-resultsProcessing:
			resProcessing = append(resProcessing, resProd)
		case <-stopChan:
			return
		}
	}

	maxPrice, maxRating := calcMaxInArray(resProcessing)

	fmt.Printf("Самый дорогой продукт:\n%+v\nс самым высоким рейтингом:\n%+v", maxPrice, maxRating)
}

func calcMaxInArray(arrProduct []product) (product, product) {
	var maxPrice product
	sort.Slice(arrProduct, func(i, j int) (less bool) {
		if arrProduct[i].Price == arrProduct[j].Price {
			return arrProduct[i].Rating >= arrProduct[j].Rating
		}
		return arrProduct[i].Price > arrProduct[j].Price
	})

	maxPrice = arrProduct[0]

	sort.Slice(arrProduct, func(i, j int) (less bool) {
		if arrProduct[i].Rating == arrProduct[j].Rating {
			return arrProduct[i].Price >= arrProduct[j].Price
		}
		return arrProduct[i].Rating > arrProduct[j].Rating
	})

	return maxPrice, arrProduct[0]
}

func fileProcessingJSON(resultsProcessing chan<- product, stopChan chan<- struct{}) {
	fileJSON, err := os.Open(nameFileJSON)
	if err != nil {
		stopChan <- struct{}{}
		log.Println("err os.Open in fileProcessingJSON: " + err.Error())
		return
	}
	defer fileJSON.Close()

	byteFileJSON, err := ioutil.ReadAll(fileJSON)
	if err != nil {
		stopChan <- struct{}{}
		log.Println("err ioutil.ReadAll in fileProcessingJSON: " + err.Error())
		return
	}

	productsFromJSON := make([]product, 0)
	if err = json.Unmarshal(byteFileJSON, &productsFromJSON); err != nil {
		stopChan <- struct{}{}
		log.Println("err json.Unmarshal in fileProcessingJSON: " + err.Error())
		return
	}

	if len(productsFromJSON) == 0 {
		stopChan <- struct{}{}
		log.Println("json file does not contain products")
		return
	}

	maxPrice, maxRating := calcMaxInArray(productsFromJSON)

	resultsProcessing <- maxPrice
	resultsProcessing <- maxRating
}

func fileProcessingCSV(resultsProcessing chan<- product, stopChan chan<- struct{}) {
	fileCSV, err := os.Open(nameFileCSV)
	if err != nil {
		stopChan <- struct{}{}
		log.Println("err os.Open in fileProcessingCSV: " + err.Error())
		return
	}
	defer fileCSV.Close()

	reader := csv.NewReader(fileCSV)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		stopChan <- struct{}{}
		log.Println("err ReadAll in fileProcessingCSV: " + err.Error())
		return
	}

	var productCSV product
	var productsFromCSV []product

	if len(rawCSVdata) < 2 {
		stopChan <- struct{}{}
		log.Println("csv file does not contain products")
		return
	}

	for _, record := range rawCSVdata[1:] {
		productCSV.Product = record[0]
		productCSV.Price, err = strconv.Atoi(record[1])
		if err != nil {
			stopChan <- struct{}{}
			log.Println("err strconv.Atoi in fileProcessingCSV 1: " + err.Error())
			return
		}
		productCSV.Rating, err = strconv.Atoi(record[2])
		if err != nil {
			stopChan <- struct{}{}
			log.Println("err strconv.Atoi in fileProcessingCSV 2: " + err.Error())
			return
		}
		productsFromCSV = append(productsFromCSV, productCSV)
	}

	maxPrice, maxRating := calcMaxInArray(productsFromCSV)

	resultsProcessing <- maxPrice
	resultsProcessing <- maxRating
}
