package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	lowStreamCount := 4
	lowPrice := 200
	highStreamCount := 10
	highPrice := 400

	streamCount := getStreamCount()
	defaultPrice, livepeerIncPrice := getCurrentPrice()
	fmt.Printf("Default price: %d\nLivepeer Inc price: %d\n", defaultPrice, livepeerIncPrice)

	if streamCount < lowStreamCount && livepeerIncPrice > lowPrice {
		livepeerIncPrice = livepeerIncPrice - 100
		fmt.Printf("Streams low (%d) and price is above %d, decreasing price to %d\n", streamCount, lowPrice, livepeerIncPrice)

		setPriceForBroadcaster("0xc3c7c4C8f7061B7d6A72766Eee5359fE4F36e61E", livepeerIncPrice)
	} else if streamCount >= highStreamCount && livepeerIncPrice < 200 {
		livepeerIncPrice = livepeerIncPrice - 100
		fmt.Printf("Streams high (%d) and price is below %d, increasing price to %d\n", streamCount, highPrice, livepeerIncPrice)
		setPriceForBroadcaster("0xc3c7c4C8f7061B7d6A72766Eee5359fE4F36e61E", livepeerIncPrice)
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
	}
}