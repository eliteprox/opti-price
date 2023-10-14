package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {

	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatal(err)
	}

	type Config struct {
		LowStreamCount    int `json:"low_stream_count"`
		HighStreamCount   int `json:"high_stream_count"`
		LowPrice          int `json:"low_price"`
		TargetStreamCount int `json:"target_stream_count"`
		HighPrice         int `json:"high_price"`
		PriceIncrement    int `json:"price_increment"`
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config: %+v\n", config)

	highStreamCount := config.HighStreamCount
	lowPrice := config.LowPrice
	targetStreamCount := config.TargetStreamCount
	highPrice := config.HighPrice
	priceIncrement := config.PriceIncrement

	streamCount := getStreamCount()
	_, livepeerIncPrice := getCurrentPrice()

	if streamCount <= highStreamCount && streamCount >= targetStreamCount {
		fmt.Printf("Streams in range %d-%d at (%d), not adjusting\n", targetStreamCount, highStreamCount, streamCount)
	} else if livepeerIncPrice-priceIncrement >= lowPrice && streamCount <= highStreamCount && streamCount < targetStreamCount {
		newPrice := livepeerIncPrice - priceIncrement
		fmt.Printf("Streams (%d) and price (%d), decreasing to %d\n", streamCount, livepeerIncPrice, newPrice)
		setPriceForBroadcaster("0xc3c7c4C8f7061B7d6A72766Eee5359fE4F36e61E", newPrice)
	} else if livepeerIncPrice <= highPrice && (streamCount >= highStreamCount || streamCount > targetStreamCount) {
		newPrice := livepeerIncPrice + priceIncrement
		fmt.Printf("Streams (%d) and price (%d), increasing to %d\n", streamCount, livepeerIncPrice, newPrice)
		setPriceForBroadcaster("0xc3c7c4C8f7061B7d6A72766Eee5359fE4F36e61E", newPrice)
	} else {
		fmt.Printf("No rules matched, streams (%d), price (%d), not adjusting\n", streamCount, livepeerIncPrice)
	}
}

func getStreamCount() int {
	url := "http://159.203.117.247:9090/api/v1/query?query=livepeer_current_sessions_total"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	d := data["data"].(map[string]interface{})
	instances := d["result"].([]interface{})

	for _, instance := range instances {
		if instance.(map[string]interface{})["metric"].(map[string]interface{})["instance"] == "68.131.51.165:80" {
			streamCount, err := strconv.Atoi(instance.(map[string]interface{})["value"].([]interface{})[1].(string))
			if err != nil {
				return 0
			} else {
				return streamCount
			}
		}
	}
	return 0
}

func getCurrentPrice() (int, int) {
	url := "http://127.0.0.1:7935/status"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	broadcasterPrices := data["BroadcasterPrices"].(map[string]interface{})
	defaultPrice, _ := strconv.Atoi(broadcasterPrices["default"].(string))
	livepeerIncPrice, _ := strconv.Atoi(broadcasterPrices["0xc3c7c4c8f7061b7d6a72766eee5359fe4f36e61e"].(string))

	return defaultPrice, livepeerIncPrice
}

func setPriceForBroadcaster(broadcaster string, price int) {
	url := "http://127.0.0.1:7935/setPriceForBroadcaster"
	data := []byte("pixelsPerUnit=1&pricePerUnit=" + strconv.Itoa(price) + "&broadcasterEthAddr=" + broadcaster)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Error setting price for %s: %d\n", broadcaster, price)
	} else {
		file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		logger := log.New(file, "set price: ", log.LstdFlags)
		logger.Printf("Price set to %d for %s\n", price, broadcaster)
	}
}
