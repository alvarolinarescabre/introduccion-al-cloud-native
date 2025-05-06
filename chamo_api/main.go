package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/gocolly/colly"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

// Health represents the response of the "get health" operation.
type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

// Link response
type Link struct {
	Id    int    `json:"id" doc:"Id of the resource"`
	Url   string `json:"url,omitempty" doc:"Url to search"`
	Links int    `json:"links" doc:"Number of the links finds"`
	Time  string `json:"time" doc:"Time take to search"`
}

// Links response
type LinksOutput struct {
	Body struct {
		Links []Link `json:"links" doc:"Links to search"`
	}
}

// 10 Websites and search https and http
var urls = []string{
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

func webScrapingCounter(url string) int {

	count := 0

	// Create a new collector
	c := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// Set a callback for when a visited HTML element is found
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

	return count
}

func getHealthCheck(api huma.API) {
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

}

func getLink(api huma.API) {

	link := make([]Link, 0)

	// Register GET /v1/link/{id}
	// This endpoint will search for link in only one of 10 websites
	huma.Register(api, huma.Operation{
		OperationID: "get-link",
		Method:      http.MethodGet,
		Path:        "/v1/link/{id}",
		Summary:     "Get link.",
		Description: "Get link to 'https' and 'http' search for one of 10 websites",
		Tags:        []string{"Links"},
	}, func(ctx context.Context, input *struct {
		Id int `path:"id" maxLength:"2" example:"0" doc:"Id of website from array"`
	}) (*LinksOutput, error) {
		resp := &LinksOutput{}

		start := time.Now()

		fmt.Println("Starting to search link...")

		id := input.Id

		// Check if the id is between 0 and 9
		if id < 0 || id > 9 {
			return nil, fmt.Errorf("id must be between 0 and 9")
		}
		// Get the url from the urls array
		url := urls[id]

		count := webScrapingCounter(url)

		timeElapsed := time.Since(start)

		link = append(link, Link{
			Id:    id,
			Url:   urls[id],
			Links: count,
			Time:  timeElapsed.String(),
		})

		resp.Body.Links = link
		fmt.Printf("id: %d | url: %s | link: %d | time: %s\n", id, urls[id], count, timeElapsed.String())

		fmt.Println("Finished searching link.")

		link = []Link{}

		return resp, nil
	})
}

func getLinks(api huma.API) {

	links := []Link{}

	// Register GET /v1/links
	// This endpoint will search for links in the 10 websites
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

		// Set a callback for when a visited HTML element is found
		for index, url := range urls {

			wg := sync.WaitGroup{}
			wg.Add(1)
			count := webScrapingCounter(url)
			wg.Done()
			wg.Wait()

			timeElapsed := time.Since(start)

			links = append(links, Link{
				Id:    index,
				Url:   url,
				Links: count,
				Time:  timeElapsed.String(),
			})
			resp.Body.Links = links
			fmt.Printf("id: %d | url: %s | links: %d | time: %s\n", index, url, count, timeElapsed.String())
		}

		fmt.Println("Finished searching links.")

		links = []Link{}

		return resp, nil
	})
}

func main() {
	// Define the options for the CLI
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("Get links to 'https' and 'http' from 10 sites.", "1.0.0"))

		// Call functions
		getHealthCheck(api)
		getLink(api)
		go getLinks(api)

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
	})

	cli.Run()
}
