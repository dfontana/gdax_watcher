package main

import (
	"log"
	"time"
	"math"

	"./common"
	gdax "github.com/preichenberger/go-gdax"
	gdaxClient "./gdax"
	pol "./poloniex"
)

func main() {
	gdaxLive()
}

func gdaxMACD() {
	daysBack := 150

	// Past 150 days for ETH daily.
	s := time.Now()
	var records []gdax.HistoricRate
	start := time.Now().AddDate(0, 0, -daysBack)
	end := time.Now()
	gran := 24*60*60
	expected := int(math.Ceil(end.Sub(start).Seconds() / float64(gran))) + 1
	for {
		records = gdaxClient.Historic("ETH-USD", start, end, gran)
		log.Printf("Data returned from API: %d/%d\n", len(records), expected)
		if len(records) == expected {
			break // Correct amount of data found
		}
		time.Sleep(time.Duration(3) * time.Second)
	}
	e1 := time.Since(s)

	s = time.Now()
	// Reduce to array of close values & their times
	hist := make([]common.TimeSeries, len(records))
	for i, val := range records {
		hist[i] = common.TimeSeries{
			val.Time,
			val.High,
		}
	}
	e2 := time.Since(s)

	// MACD: 12 fast, 26 slow, 9 signal
	s = time.Now()
	comp := common.MACD{}
	if err := comp.Populate(hist, 12, 26, 9); err != nil {
		log.Fatal(err)
	}
	comp.Plot("./test.png")
	e3 := time.Since(s)

	log.Printf("Timings:\n\tHistory: %s\n\tTimeSeries: %s\n\tMACD: %s", e1,e2,e3)
}

func polHist() {
	p := pol.Poloniex{false, []string{"USDT_ETH"}}
	recsP := p.Historic("USDT_ETH", time.Date(2017, time.December, 14, 0, 0, 0, 0, time.Local), time.Now())
	p.CSV("./outP.csv", recsP)
}

func gdaxHist() {
	recs := gdaxClient.Historic("ETH-USD", time.Date(2017, time.December, 14, 0, 0, 0, 0, time.Local), time.Now(), 200)
	gdaxClient.CSV("./outG.csv", recs)
}

func polLive() {
	p := pol.Poloniex{true, []string{"USDT_ETH"}}
	go p.Live()
	time.Sleep(10 * time.Second)
	p.Enabled = false
}

func gdaxLive() {
	// Asynchronously fetch data to messages channel.
	messages := make(chan gdaxClient.WsMatch)
	quit := make(chan bool)
	go gdaxClient.Live([]string{"ETH-USD", "BTC-USD"}, messages, quit)

	// Kill the livefeed after 10 seconds.
	go func() {
		time.Sleep(10 * time.Second)
		quit <- true
	}()

	// Loop until something stops the socket feed (error or disabled)
	for msg := range messages {
		log.Printf("%+v\n", msg)
	}
}
