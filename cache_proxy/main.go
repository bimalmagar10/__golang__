package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)



type CacheEntry struct {
	Body []byte
	Headers http.Header
	StatusCode int
	Expires time.Time
}

type Cache struct {
	entries map[string]*CacheEntry
	mu sync.RWMutex
}

var (
	port int
	origin string
	clearCache bool
	cache = NewCache()
)

func NewCache() *Cache {
	return &Cache {
		entries: make(map[string]*CacheEntry),
	}
}

func (c *Cache) Set(key string,entry *CacheEntry){
	// c.mu.Lock()
	// defer c.mu.Unlock()
	c.entries[key] = entry
	fmt.Println()
}

func (c *Cache) Get(key string) (*CacheEntry,bool){
	// c.mu.RLock()
	// defer c.mu.RUnlock()
	entry,exists := c.entries[key]

	if !exists || time.Now().After(entry.Expires) {
		return nil,false
	}

	return entry,true
}


func main() {
	fmt.Println(strings.Repeat("=",80))
	fmt.Println("Welcome to Bimal's simple cache proxy in Golang!")
	fmt.Println(strings.Repeat("=",80))

	//Flags definition
	flag.IntVar(&port,"port",0,"Port for proxy Server")
	flag.StringVar(&origin,"origin","","URL on which url to proxy for")
	flag.BoolVar(&clearCache,"clear-cache",false,"Clearing the cache")

	//This will parse all the arguments passed as a flag in CLI
	flag.Parse()

	if !isValidPort(port) || origin == "" {
		log.Fatal("Both --port and --origin are required!")
	}

	if !isValidURL(origin) {
		log.Fatal(("Invalid URL"))
	}

	StartProxyServer()
	

}

func  isValidPort(port int) bool  {
	if reflect.TypeOf(port).Kind() == reflect.Int {
		return port > 0 && port <=65535
	}
	return false
}

func isValidURL(url string) bool {
	urlRgx := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|\/|\/\/)?[A-z0-9_-]*?[:]?[A-z0-9_-]*?[@]?[A-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
	return urlRgx.MatchString(url)
}

func StartProxyServer() {
	originURL,err := url.Parse(origin)
	
	if err != nil {
		log.Fatalf("Invalid origin URL: %v",err)
	}

	proxy := httputil.NewSingleHostReverseProxy(originURL)
	proxy.Director = ModifyRequestDirector(originURL)

	proxy.ModifyResponse = ModifyResponseHandler(originURL)

	http.HandleFunc("/",RequestHandler(proxy,originURL))
	log.Printf("Starting proxy server on port %d for origin %s \n",port,origin)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",port),nil))

}

func ModifyRequestDirector(originURL *url.URL) func(req *http.Request) {
	return func (req *http.Request) {
		req.URL.Scheme = originURL.Scheme
		req.URL.Host = originURL.Host
		req.Host = originURL.Host
		req.Header.Set("X-Forwarded-Host",req.Host)
	}
}

func ModifyResponseHandler(originURL *url.URL) func(*http.Response) error {
	return func (resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(body))

		headers := make(http.Header)

		for k, v := range resp.Header {
			headers[k] = v
		}

		entry := &CacheEntry{
			Body:       body,
			Headers:    headers,
			StatusCode: resp.StatusCode,
			Expires:    time.Now().Add(5 * time.Minute),
		}

		key := CacheKey(resp.Request,originURL.Host)

		cache.Set(key, entry)
		resp.Header.Set("X-Cache","MISS")
		return nil
	}
}

func CacheKey(r *http.Request,originHost string) string {
	return fmt.Sprintf("%s %s%s", 
        r.Method, 
        originHost, 
        r.URL.RequestURI()) 
}

func RequestHandler(proxy *httputil.ReverseProxy,originURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request) {
		key := CacheKey(r,originURL.Host)
		if entry,hit := cache.Get(key); hit {
			if entry.StatusCode == 0 || len(entry.Body) == 0 {
				log.Printf("Invalid cache entry for %s", key)
				proxy.ServeHTTP(w, r)
				return
			}
			SendCachedResponse(w,entry)
			return
		}

		// If Cache Miss - Forward back to origin
		proxy.ServeHTTP(w,r)
	}
}

func SendCachedResponse(w http.ResponseWriter,entry *CacheEntry) {

	for k,v := range entry.Headers {
		w.Header()[k] = v
	}

	fmt.Printf("%v",w.Header())

	w.Header().Set("X-Cache","HIT")
	w.WriteHeader(entry.StatusCode)
	w.Write(entry.Body)
}