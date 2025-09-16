package scanner

import (
	"context"
	"fmt"
	db "go-deadlink-scanner/internal/database/sqlc"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

type Service struct {
	queries    *db.Queries
	visited    map[string]bool
	mu         sync.Mutex
	wg         sync.WaitGroup
	results    chan Result
	sem        chan struct{}
	MaxWorkers int
}

type Result struct {
	Link   string
	Status string
}

func NewService(queries *db.Queries, maxWorkers int) *Service {
	return &Service{
		queries:    queries,
		visited:    make(map[string]bool),
		results:    make(chan Result, 100),
		sem:        make(chan struct{}, maxWorkers),
		MaxWorkers: maxWorkers,
	}
}

func (s *Service) Scan(startUrl string, userId int32) ([]db.Result, error) {
	s.visited = make(map[string]bool)
	s.results = make(chan Result, 100)

	rawResults := s.startScan(startUrl)

	unique := make(map[string]bool)
	stored := make([]db.Result, 0, len(rawResults))
	for _, r := range rawResults {
		if r.Link == "" {
			continue
		}
		if unique[r.Link] {
			continue
		}
		unique[r.Link] = true
		res, err := s.queries.CreateResult(context.Background(), db.CreateResultParams{
			UserID:  userId,
			PageUrl: startUrl,
			LinkUrl: r.Link,
			Status:  r.Status,
		})
		if err != nil {
			return stored, err
		}
		stored = append(stored, res)
	}
	return stored, nil
}

func (s *Service) startScan(startURL string) []Result {
	s.wg.Add(1)
	s.acquire()
	go func() {
		defer s.release()
		s.scanURL(startURL, startURL)
	}()

	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	var res []Result
	for r := range s.results {
		res = append(res, r)
	}

	return res
}

func (s *Service) scanURL(baseURL, currentURL string) {
	defer s.wg.Done()

	s.mu.Lock()
	if s.visited[currentURL] {
		s.mu.Unlock()
		return
	}
	s.visited[currentURL] = true
	s.mu.Unlock()

	links, err := s.fetchPageLinks(currentURL)
	if err != nil {
		s.results <- Result{Link: currentURL, Status: "Error fetching page"}
		return
	}

	for _, link := range links {
		abs := s.normalizeURL(baseURL, link)
		if abs == "" {
			continue
		}

		if sameDomain(baseURL, abs) {
			s.wg.Add(1)
			s.acquire()
			go func(l string) {
				defer s.release()
				defer s.wg.Done()
				status := s.checkLink(l)
				s.results <- Result{Link: l, Status: status}
			}(abs)

			s.mu.Lock()
			if !s.visited[abs] {
				s.visited[abs] = true
				s.wg.Add(1)
				s.acquire()
				go func(next string) {
					defer s.release()
					s.scanURL(baseURL, next)
				}(abs)
			}
			s.mu.Unlock()
		}
	}
}

func (s *Service) checkLink(link string) string {
	resp, err := http.Head(link)
	if err != nil {
		respGet, err := http.Get(link)
		if err != nil {
			return "Dead"
		}
		defer respGet.Body.Close()
		if respGet.StatusCode >= 200 && respGet.StatusCode < 400 {
			return "Alive"
		}
		return fmt.Sprintf("Code %d", respGet.StatusCode)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "Alive"
	}
	return fmt.Sprintf("Code %d", resp.StatusCode)
}

func (s *Service) fetchPageLinks(pageURL string) ([]string, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var links []string
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

func sameDomain(startURL, checkURL string) bool {
	start, err := url.Parse(startURL)
	if err != nil {
		return false
	}
	check, err := url.Parse(checkURL)
	if err != nil {
		return false
	}
	return start.Hostname() == check.Hostname()
}

func (s *Service) normalizeURL(base, href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if u.IsAbs() {
		return u.String()
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}

func (s *Service) acquire() {
	s.sem <- struct{}{}
}

func (s *Service) release() {
	<-s.sem
}
