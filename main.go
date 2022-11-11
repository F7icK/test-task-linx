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
	resultsProcessing := make(chan product, 2)
	stopChan := make(chan struct{}, 1)

	go fileProcessingJSON(resultsProcessing, stopChan)
	go fileProcessingCSV(resultsProcessing, stopChan)

	var resProducts []product
	for i := 0; i < 2; i++ {
		select {
		case resProd := <-resultsProcessing:
			resProducts = append(resProducts, resProd)
		case <-stopChan:
			return
		}
	}

	// Если я правильно понял из условия задачи, приоритет сортировки в цене продукта.
	sort.Slice(resProducts, func(i, j int) (less bool) {
		return resProducts[i].Rating >= resProducts[j].Rating && resProducts[i].Price >= resProducts[j].Price
	})

	fmt.Println(resProducts[0])
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

	sort.Slice(productsFromJSON, func(i, j int) (less bool) {
		return productsFromJSON[i].Rating >= productsFromJSON[j].Rating && productsFromJSON[i].Price >= productsFromJSON[j].Price
	})

	resultsProcessing <- productsFromJSON[0]
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

	sort.Slice(productsFromCSV, func(i, j int) (less bool) {
		return productsFromCSV[i].Rating >= productsFromCSV[j].Rating && productsFromCSV[i].Price >= productsFromCSV[j].Price
	})

	resultsProcessing <- productsFromCSV[0]
}
