package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// --- Models ---

type WSMessage struct {
	Action string `json:"action"`
	Arg    struct {
		InstType string `json:"instType"`
		Channel  string `json:"channel"`
		InstId   string `json:"instId"`
	} `json:"arg"`
	Data []struct {
		Asks [][]string `json:"asks"` // WebSocket returns strings: ["price", "size"]
		Bids [][]string `json:"bids"`
		Ts   string     `json:"ts"`
	} `json:"data"`
}

type Level struct {
	Price float64
	Size  float64
}

func main() {
	// 1. Setup Signal Handling for clean exit (Effective Go style)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 2. Define Connection
	u := url.URL{Scheme: "wss", Host: "ws.bitget.com", Path: "/v2/ws/public"}
	fmt.Printf("Connecting to %s...\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// 3. Subscribe to Order Book (books15 provides a constant stream of snapshots)
	subRequest := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"instType": "USDT-FUTURES",
				"channel":  "books15",
				"instId":   "BTCUSDT",
			},
		},
	}
	if err := c.WriteJSON(subRequest); err != nil {
		log.Fatal("subscribe:", err)
	}

	// 4. Start Heartbeat Routine (Bitget requires "ping" string every 30s)
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := c.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
				return
			}
		}
	}()

	fmt.Println("Analysis Started. Watching for Liquidity Peaks (SD > 2.0)...")

	// 5. Main Loop: Read and Analyze
	for {
		select {
		case <-interrupt:
			fmt.Println("Interrupted. Closing connection...")
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}

			// Bitget sends "pong" as a string, ignore it
			if string(message) == "pong" {
				continue
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				continue
			}

			// Only process messages containing data
			if len(wsMsg.Data) > 0 {
				processDepth(wsMsg.Data[0].Asks, wsMsg.Data[0].Bids)
			}
		}
	}
}

// processDepth handles the conversion and statistical analysis
func processDepth(rawAsks, rawBids [][]string) {
	asks := parseLevels(rawAsks)
	bids := parseLevels(rawBids)

	// Detect peaks using a 2.0 Standard Deviation threshold
	askPeaks := detectPeaks(asks, 2.0)
	bidPeaks := detectPeaks(bids, 2.0)

	// Clear terminal effect (optional) or just print findings
	if len(askPeaks) > 0 || len(bidPeaks) > 0 {
		fmt.Printf("\n[%s] --- Peak Detected ---\n", time.Now().Format("15:04:05"))
		for _, p := range askPeaks {
			fmt.Printf("RESISTANCE | Price: %.1f | Vol: %.3f (Peak)\n", p.Price, p.Size)
		}
		for _, p := range bidPeaks {
			fmt.Printf("SUPPORT    | Price: %.1f | Vol: %.3f (Peak)\n", p.Price, p.Size)
		}
	}
}

func parseLevels(raw [][]string) []Level {
	levels := make([]Level, 0, len(raw))
	for _, item := range raw {
		if len(item) < 2 {
			continue
		}
		p, _ := strconv.ParseFloat(item[0], 64)
		s, _ := strconv.ParseFloat(item[1], 64)
		levels = append(levels, Level{Price: p, Size: s})
	}
	return levels
}

func detectPeaks(levels []Level, threshold float64) []Level {
	if len(levels) == 0 {
		return nil
	}

	var sum, sumSq float64
	for _, l := range levels {
		sum += l.Size
		sumSq += l.Size * l.Size
	}

	n := float64(len(levels))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	stdDev := math.Sqrt(variance)
	limit := mean + (threshold * stdDev)

	var peaks []Level
	for _, l := range levels {
		// Effective Go: Filtering logic
		if l.Size > limit && l.Size > 0.5 { // Added a minimum floor of 0.5 BTC to avoid noise
			peaks = append(peaks, l)
		}
	}
	return peaks
}
