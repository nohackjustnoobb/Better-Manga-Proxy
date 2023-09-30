package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

func countCacheSize() float64 {
	defer func() {
		if recover() != nil {
			return
		}
	}()

	var dirSize int64 = 0

	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSize += file.Size()
		}

		return nil
	}

	filepath.Walk("cache", readSize)
	sizeMB := float64(dirSize) / 1024.0 / 1024.0

	return sizeMB
}

func cacheManager() {
	defer fmt.Println("* CacheManager disabled *")

	maxCacheSizeStr := os.Getenv("MAX_CACHE_SIZE")
	if maxCacheSizeStr == "" {
		return
	}

	maxCacheSize, err := strconv.ParseFloat(maxCacheSizeStr, 64)
	if err != nil {
		return
	}

	for {
		cacheSize := countCacheSize()

		if cacheSize > maxCacheSize {
			// save the file info
			var fileInfo []os.FileInfo

			saveInfo := func(path string, file os.FileInfo, err error) error {
				if !file.IsDir() {
					fileInfo = append(fileInfo, file)
				}

				return nil
			}

			filepath.Walk("cache", saveInfo)

			// sort the files by modification time
			sort.SliceStable(fileInfo, func(i, j int) bool {
				return fileInfo[i].ModTime().After(fileInfo[j].ModTime())
			})

			// save the name of the file needed to be deleted
			var fileShouldDeleted []string
			var overflowMB = cacheSize - maxCacheSize

			for _, file := range fileInfo {
				if overflowMB <= 0 {
					break
				}

				fileShouldDeleted = append(fileShouldDeleted, file.Name())
				overflowMB -= float64(file.Size()) / 1024.0 / 1024.0
			}

			for _, name := range fileShouldDeleted {
				os.Remove("cache/" + name)
			}
		}

		// check cache size every 1 minute
		time.Sleep(time.Minute)
	}

}
