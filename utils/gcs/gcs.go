package gcsutils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/ct-logic-api-document/internal/constants"
	"golang.org/x/xerrors"
)

var ErrDone = errors.New("no more items in iterator")

func GetFolderPath(ctx context.Context, startTime, endTime time.Time) ([]string, error) {
	urls := []string{}
	// load nothing
	if startTime.After(endTime) {
		return urls, nil
	}
	curr := startTime
	for {
		hourFolder := curr.Format("2006-01-02/15")
		// form the gcs url path
		url := fmt.Sprintf("logs/%s/%s/", constants.ServiceNameProxy, hourFolder)
		urls = append(urls, url)
		curr = curr.Add(time.Hour)
		if curr.After(endTime) {
			break
		}
	}
	return urls, nil
}

func ListFilesByFolders(ctx context.Context, bucket *storage.BucketHandle, urls []string) (map[string][]string, error) {
	files := map[string][]string{}
	for _, url := range urls {
		it := bucket.Objects(ctx, &storage.Query{
			Prefix: url,
		})
		for {
			obj, err := it.Next()
			if xerrors.Is(err, ErrDone) || obj == nil {
				break
			}
			if err != nil {
				log.Printf("listBucket: unable to list bucket %q: %v", constants.LogArchivalBucket, err)
				return nil, err
			}
			if strings.HasSuffix(obj.Name, ".metadata") {
				continue
			}
			if !strings.HasSuffix(obj.Name, ".gz") {
				continue
			}

			_, ok := files[url]
			if ok {
				files[url] = append(files[url], obj.Name)
			} else {
				files[url] = []string{obj.Name}
			}

		}
	}
	return files, nil
}
