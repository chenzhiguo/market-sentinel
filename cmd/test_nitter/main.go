package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	hosts := []string{
		"https://nitter.poast.org",
		"https://nitter.privacydev.net",
		"https://nitter.lucabased.xyz",
		"https://nitter.moomoo.me",
		"https://nitter.soopy.moe",
		"https://xcancel.com",
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Println("Testing Nitter instances...")
	for _, host := range hosts {
		start := time.Now()
		resp, err := client.Get(host + "/elonmusk/rss")
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ %s: Error (%v)\n", host, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Printf("✅ %s: OK (%dms)\n", host, duration.Milliseconds())
		} else {
			fmt.Printf("❌ %s: Status %d\n", host, resp.StatusCode)
		}
	}
}
