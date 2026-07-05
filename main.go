package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CryptoPrice struct {
	USD float64 `json:"usd"`
}

var (
	cryptoPriceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_price_usd",
			Help: "Current price of cryptocurrencies in USD",
		},
		[]string{"coin"},
	)
)

func main() {

	prometheus.MustRegister(cryptoPriceGauge)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	var get_price = "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum,solana&vs_currencies=usd"

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()

	for range ticker.C {
		// Perform the GET request to the CoinGecko API
		resp, err := http.Get(get_price)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			log.Fatalf("Request failed: %v", err)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}
		resp.Body.Close()

		var data map[string]CryptoPrice
		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		for coinName, priceInfo := range data {
			fmt.Println(coinName, ":", priceInfo.USD)
			cryptoPriceGauge.WithLabelValues(coinName).Set(priceInfo.USD)
		}
	}
}

/*

CoinGecko API - Keyless: https://docs.coingecko.com/docs/keyless-public-api
curl -X GET \
  "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum,solana&vs_currencies=usd"

*/
