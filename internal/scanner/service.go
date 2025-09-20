package scanner

import (
	"context"
	"fmt"
	db "go-deadlink-scanner/internal/database/sqlc"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html"
)

type Service struct {
	queries    *db.Queries
	maxWorkers int
	client     *http.Client

	results      map[string]*ScanResult
	resultsMutex sync.Mutex
	visited      map[string]bool
	visitedMutex sync.Mutex
}

type ScanResult struct {
	URL        string
	Status     string
	StatusCode int
	Error      string
}

type linkJob struct {
	url     string
	baseURL *url.URL
	depth   int
}

func NewService(queries *db.Queries, maxWorkers int) *Service {
	return &Service{
		queries:    queries,
		maxWorkers: maxWorkers,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("stopped after 5 redirects")
				}
				return nil
			},
		},
		results: make(map[string]*ScanResult),
		visited: make(map[string]bool),
	}
}

func (s *Service) Scan(startURL string, userID int32) ([]db.Result, error) {
	s.resultsMutex.Lock()
	s.results = make(map[string]*ScanResult)
	s.resultsMutex.Unlock()

	s.visitedMutex.Lock()
	s.visited = make(map[string]bool)
	s.visitedMutex.Unlock()

	baseURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	jobs := make(chan linkJob, 1000)
	var wg sync.WaitGroup
	var activeJobs int32

	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go s.workerWithJobTracking(i, jobs, &wg, &activeJobs)
	}

	atomic.AddInt32(&activeJobs, 1)
	jobs <- linkJob{url: startURL, baseURL: baseURL, depth: 0}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			time.Sleep(100 * time.Millisecond)
			if atomic.LoadInt32(&activeJobs) == 0 {
				close(jobs)
				wg.Wait()
				return
			}
		}
	}()

	select {
	case <-done:
		log.Printf("Scan completed. Found %d links", len(s.results))
	case <-time.After(30 * time.Second):
		log.Printf("Scan timed out after 30 seconds. Found %d links so far", len(s.results))
		close(jobs)
		wg.Wait()
	}

	var dbResults []db.Result
	s.resultsMutex.Lock()
	for url, result := range s.results {
		dbResult := db.Result{
			UserID:    userID,
			PageUrl:   startURL,
			LinkUrl:   url,
			Status:    result.Status,
			CheckedAt: time.Now(),
		}
		dbResults = append(dbResults, dbResult)

		if s.queries != nil {
			_, err := s.queries.CreateResult(context.Background(), db.CreateResultParams{
				UserID:  userID,
				PageUrl: startURL,
				LinkUrl: url,
				Status:  result.Status,
			})
			if err != nil {
				log.Printf("Failed to save result for %s: %v", url, err)
			}
		}
	}
	s.resultsMutex.Unlock()

	return dbResults, nil
}

func (s *Service) workerWithJobTracking(id int, jobs chan linkJob, wg *sync.WaitGroup, activeJobs *int32) {
	defer wg.Done()

	for job := range jobs {
		s.visitedMutex.Lock()
		if s.visited[job.url] {
			s.visitedMutex.Unlock()
			atomic.AddInt32(activeJobs, -1)
			continue
		}
		s.visited[job.url] = true
		s.visitedMutex.Unlock()

		log.Printf("Worker %d: checking %s (depth: %d)", id, job.url, job.depth)

		result := s.checkLink(job.url)

		s.resultsMutex.Lock()
		s.results[job.url] = result
		s.resultsMutex.Unlock()

		if result.StatusCode == 200 && job.depth < 10 && strings.Contains(result.Status, "text/html") {
			links := s.extractLinks(job.url, job.baseURL)

			newJobsAdded := 0
			for _, link := range links {
				s.visitedMutex.Lock()
				if !s.visited[link] {
					select {
					case jobs <- linkJob{url: link, baseURL: job.baseURL, depth: job.depth + 1}:
						newJobsAdded++
					default:
						// Channel is full, skip this link
					}
				}
				s.visitedMutex.Unlock()
			}

			if newJobsAdded > 0 {
				atomic.AddInt32(activeJobs, int32(newJobsAdded))
			}
		}

		atomic.AddInt32(activeJobs, -1)
	}
}

func (s *Service) checkLink(linkURL string) *ScanResult {
	result := &ScanResult{
		URL:    linkURL,
		Status: "unknown",
	}

	req, err := http.NewRequest("HEAD", linkURL, nil)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("invalid request: %v", err)
		return result
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; DeadLinkChecker/1.0)")

	resp, err := s.client.Do(req)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		result.Status = fmt.Sprintf("ok (%s)", resp.Header.Get("Content-Type"))
	case resp.StatusCode == 404:
		result.Status = "404 Not Found"
		result.Error = "Page not found"
	case resp.StatusCode >= 500:
		result.Status = fmt.Sprintf("Server Error (%d)", resp.StatusCode)
		result.Error = "Internal server error"
	case resp.StatusCode >= 400:
		result.Status = fmt.Sprintf("Client Error (%d)", resp.StatusCode)
		result.Error = "Client error"
	default:
		result.Status = fmt.Sprintf("Redirect (%d)", resp.StatusCode)
	}

	return result
}

func (s *Service) extractLinks(pageURL string, baseURL *url.URL) []string {
	resp, err := s.client.Get(pageURL)
	if err != nil {
		log.Printf("Failed to get page %s: %v", pageURL, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		log.Printf("Failed to read body from %s: %v", pageURL, err)
		return nil
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Failed to parse HTML from %s: %v", pageURL, err)
		return nil
	}

	var links []string
	s.traverseHTML(doc, baseURL, &links)

	linkMap := make(map[string]bool)
	var uniqueLinks []string
	for _, link := range links {
		if !linkMap[link] {
			linkMap[link] = true
			uniqueLinks = append(uniqueLinks, link)
		}
	}

	return uniqueLinks
}

func (s *Service) traverseHTML(n *html.Node, baseURL *url.URL, links *[]string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				link := s.resolveURL(attr.Val, baseURL)
				if link != "" && s.isInternalLink(link, baseURL) {
					*links = append(*links, link)
				}
				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.traverseHTML(c, baseURL, links)
	}
}

func (s *Service) resolveURL(href string, baseURL *url.URL) string {
	if href == "" || strings.HasPrefix(href, "#") ||
		strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
		return ""
	}

	linkURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolvedURL := baseURL.ResolveReference(linkURL)
	return resolvedURL.String()
}

func (s *Service) isInternalLink(link string, baseURL *url.URL) bool {
	linkURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	return linkURL.Host == baseURL.Host
}
