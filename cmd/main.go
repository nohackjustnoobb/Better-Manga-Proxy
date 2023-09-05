package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/inancgumus/screen"
	"github.com/joho/godotenv"
	"github.com/sunshineplan/imgconv"
	"github.com/valyala/fasthttp"
	fastprefork "github.com/valyala/fasthttp/prefork"
)

var client *fasthttp.Client
var settings map[string]any

var maxClientEachRequest = 3

type Image struct {
	hash string
	data []byte
}

var imagePool = sync.Pool{
	New: func() interface{} { return new(Image) },
}

func shuffle(array []string) {
	for i := range array {
		j := rand.Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
}

func fetchImage(driver string, destination string, genre string) (contentType string, body []byte) {
	// generate all the urls
	var urls []string
	desUrl, _ := url.Parse(destination)

	for _, host := range settings[driver].(map[string]any)["genre"].(map[string]any)[genre].([]interface{}) {
		u, _ := url.Parse(host.(string))
		u.Path = desUrl.Path
		u.RawQuery = desUrl.RawQuery

		urls = append(urls, u.String())
	}

	if len(urls) == 0 {
		urls = append(urls, destination)
	}

	if len(urls) > maxClientEachRequest {
		shuffle(urls)
		urls = urls[:maxClientEachRequest]
	}

	// fetch the image
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(1)

	for _, url := range urls {
		go func(url string) {

			req := fasthttp.AcquireRequest()
			req.Header.SetMethod(fasthttp.MethodGet)

			// set headers
			for key, value := range settings[driver].(map[string]any)["headers"].(map[string]interface{}) {
				req.Header.Add(key, value.(string))
			}

			// send the request
			req.SetRequestURI(url)
			resp := fasthttp.AcquireResponse()
			err := client.Do(req, resp)
			fasthttp.ReleaseRequest(req)

			if err == nil {
				contentType = string(resp.Header.ContentType())
				body = resp.Body()

				// lock the mutex
				if mu.TryLock() {
					wg.Done()
				} else {
					return
				}

			}
			fasthttp.ReleaseResponse(resp)

		}(url)
	}

	wg.Wait()

	return contentType, body
}

func saveImage(image *Image) {
	defer imagePool.Put(image)

	defer func() {
		if recover() != nil {
			os.Remove("cache/" + image.hash + ".jpeg")
		}
	}()

	// create dir if it doesn't exist
	newpath := filepath.Join(".", "cache")
	os.MkdirAll(newpath, os.ModePerm)

	fo, _ := os.Create("cache/" + image.hash + ".jpeg")
	src, _ := imgconv.Decode(bytes.NewReader(image.data))
	imgconv.Write(fo, src, &imgconv.FormatOption{Format: imgconv.JPEG})
}

func mainHandler(ctx *fasthttp.RequestCtx) {
	defer log(ctx)

	defer func() {
		if recover() != nil {
			ctx.Response.SetStatusCode(400)
		}
	}()

	var driver = string(ctx.QueryArgs().Peek("driver"))
	var destination = string(ctx.QueryArgs().Peek("destination"))
	var genre = string(ctx.QueryArgs().Peek("genre"))

	if driver == "" || destination == "" || genre == "" {
		panic(nil)
	}

	// generate the hash string
	desUrl, _ := url.Parse(destination)
	h := md5.New()
	h.Write([]byte(driver + desUrl.EscapedPath()))
	hash := hex.EncodeToString(h.Sum(nil))

	// check if cached
	dat, err := os.ReadFile("cache/" + hash + ".jpeg")
	if err == nil {
		ctx.Response.Header.SetContentType("image/jpeg")
		ctx.Response.SetBody(dat)
		go updateTimestamp(hash)
		return
	}

	// fetch the image
	contentType, body := fetchImage(driver, destination, genre)

	// get image object
	var image = imagePool.Get().(*Image)
	// set data
	image.hash = hash
	image.data = body

	ctx.Response.Header.SetContentType(contentType)
	ctx.Response.SetBody(body)

	// cache the image
	go saveImage(image)
}

func main() {
	port := ":8080"
	godotenv.Load()

	initClient()
	fetchSettings()

	// Clear the screen

	if !fastprefork.IsChild() {
		screen.Clear()
		screen.MoveTopLeft()

		fmt.Println(time.Now().Format("January 2, 2006 - 15:04:05"))
		fmt.Printf("Serving at port %s \n", port)

		// Create cache manager
		go cacheManager()

	} else {
		fmt.Println("Starting worker")
	}

	server := &fasthttp.Server{
		Handler:               mainHandler,
		MaxIdleWorkerDuration: time.Minute * 30,
	}

	fastprefork.New(server).ListenAndServe(port)
}
