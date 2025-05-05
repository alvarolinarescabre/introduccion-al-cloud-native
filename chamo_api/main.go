package main

import (
	"context"
	"fmt"
	"time"
	"strings"
	"github.com/gocolly/colly"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"net/http"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// Health represents the response of the "get health" operation.
type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

// Link response
type Link struct {
	Id      int    `json:"id" doc:"Id of the resource"`
	Url     string `json:"url,omitempty" doc:"Url to search"`
	Links	int	   `json:"links" doc:"Number of the links finds"`
	Time    string `json:"time" doc:"Time take to search"`
}

// Links response
type LinksOutput struct {
	Body struct {
		Links []Link `json:"links" doc:"Links to search"`      
	}
}

func main() {
	// 10 Websites and search https and http
	urls := []string{
		"https://www.holachamo.com",
		"https://www.paradigmadigital.com",
		"https://www.realpython.com",
		"https://www.lapatilla.com",
		"https://www.facebook.com",
		"https://www.gitlab.com",
		"https://www.youtube.com",
		"https://www.mozilla.org",
		"https://www.github.com",
		"https://www.google.com",
	}

	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Get links to 'https' and 'http' from 10 sites.", "1.0.0"))

	// Add GET / for health checks
	huma.Register(api, huma.Operation{
		OperationID:   "get-health",
		Summary:       "Get health",
		Method:        http.MethodGet,
		Path:          "/",
		DefaultStatus: http.StatusOK,
	}, func(ctx context.Context, i *struct{}) (*HealthOutput, error) {
		resp := &HealthOutput{}
		resp.Body.Status = "ok"
		return resp, nil
	})

	links := make([]Link, 0)

	// Register GET /v1/tags
	huma.Register(api, huma.Operation{
		OperationID: "get-links",
		Method:      http.MethodGet,
		Path:        "/v1/links",
		Summary:     "Get links.",
		Description: "Get links to 'https' and 'http' from 10 sites.",
		Tags:        []string{"Links"},
	}, func(ctx context.Context, input *struct{}) (*LinksOutput, error) {
		resp := &LinksOutput{}

		start := time.Now()

		fmt.Println("Starting to search links...")

		// Create a new collector
		c := colly.NewCollector(
			colly.MaxDepth(10),
			colly.Async(true),
		)
		
		
		c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

		// Set a callback for when a visited HTML element is found
		for index, url := range urls {
			count := 0
			
			c.OnHTML("a[href]", func(e *colly.HTMLElement) {
				link := e.Attr("href")
				if link == "" {
					return
				} else {
					count += strings.Count(link, "https://")
					count += strings.Count(link, "http://")
				}
			})
	
	
			c.Visit(url)
			c.Wait()
			
			timeElapsed := time.Since(start)

			links = append(links, Link{
				Id:      index,
				Url:     url,
				Links:	 count,
				Time:    timeElapsed.String(),
			})
			resp.Body.Links = links
			fmt.Printf("id: %d | url: %s |  links: %d | time: %s\n", index, url, count, timeElapsed.String())
		}

		fmt.Println("Finished searching links.")

		return resp, nil
	})

	// Start the server!
	http.ListenAndServe("0.0.0.0:8888", router)
}
