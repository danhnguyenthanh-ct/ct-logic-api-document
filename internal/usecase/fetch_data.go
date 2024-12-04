package usecase

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/storage"
	"github.com/carousell/ct-go/pkg/container"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/carousell/ct-go/pkg/workerpool"
	"github.com/ct-logic-api-document/config"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	"github.com/ct-logic-api-document/internal/repository/mongodb"
	gcsutils "github.com/ct-logic-api-document/utils/gcs"
	"github.com/goccy/go-json"
	"github.com/spf13/cast"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const maxLogLinesProccessing = 100

var jsonRegexp = regexp.MustCompile(`^\{.*\}`)
var k8sPrefixRegexp = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z) (stderr|stdout) F\s+`)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024*1024) // 1MB buffer
	},
}

type IFetchDataUC interface {
	FetchDataFromGcs(ctx context.Context) error

	// testing
	ProcessLogLine(ctx context.Context, logLine []byte) error
}

type fetchDataUC struct {
	conf    *config.Config
	storage mongodb.MongoStorage
}

func NewFetchDataUC(
	conf *config.Config,
	storage mongodb.MongoStorage,
) IFetchDataUC {
	return &fetchDataUC{
		conf:    conf,
		storage: storage,
	}
}

func (f *fetchDataUC) FetchDataFromGcs(ctx context.Context) error {
	// init gcs client
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logctx.Errorw(ctx, "failed to init client", "err", err)
		return err
	}
	defer client.Close()
	// check if bucket exists
	bucket := client.Bucket(constants.LogArchivalBucket)
	exists, err := bucket.Attrs(ctx)
	if err != nil {
		logctx.Errorw(ctx, "failed to connect to bucket", "err", err)
		return err
	}
	logctx.Infof(ctx, "Bucket %s exists with metadata %+v\n", constants.LogArchivalBucket, exists)
	// get folder path
	startTime := time.Unix(1733108400, 0) // 2024-12-02 03:00:00
	endTime := time.Unix(1733112000, 0)   // 2024-12-02 04:00:00
	urls, err := gcsutils.GetFolderPath(ctx, startTime, endTime)
	if err != nil {
		logctx.Errorw(ctx, "failed to get folder path", "err", err)
		return err
	}
	if len(urls) == 0 {
		logctx.Error(ctx, "no folder path found")
		return nil
	}
	mapUrlWithFiles, err := gcsutils.ListFilesByFolders(ctx, bucket, urls)
	if err != nil {
		logctx.Errorw(ctx, "failed to list files by folders", "err", err)
		return err
	}
	if len(mapUrlWithFiles) == 0 {
		logctx.Error(ctx, "no files found")
		return nil
	}
	// create the worker pool with the number of urls
	numberOfMaxWorkers := len(mapUrlWithFiles)
	pool := workerpool.NewE(numberOfMaxWorkers)

	// wait for all tasks to finish and close the pool
	defer pool.Close()

	for _, files := range mapUrlWithFiles {
		files := files
		pool.Run(func() error {
			return f.processFolder(ctx, bucket, files)
		})
	}

	// Wait for all tasks to finish and check errors
	if err := pool.Wait(); err != nil {
		logctx.Errorw(ctx, "err when process  folder", "err", err)
	}

	return nil
}

func (f *fetchDataUC) processFolder(ctx context.Context, bucket *storage.BucketHandle, files []string) error {
	numberOfMaxWorkers := len(files) / 10
	pool := workerpool.NewE(numberOfMaxWorkers)
	// wait for all tasks to finish and close the pool
	defer pool.Close()
	for _, file := range files {
		file := file
		pool.Run(func() error {
			// process the file
			return f.processFile(ctx, bucket, file)
		})
	}

	if err := pool.Wait(); err != nil {
		logctx.Errorw(ctx, "err when process file", "err", err)
		return err
	}

	return nil
}

func (f *fetchDataUC) processFile(ctx context.Context, bucket *storage.BucketHandle, file string) error {
	// process the file
	logctx.Infof(ctx, "processing file %s", file)
	logLineChan := make(chan []byte, maxLogLinesProccessing)
	f.readFile(ctx, bucket, file, logLineChan)
	pool := workerpool.NewE(maxLogLinesProccessing)
	defer func() {
		// close the jobs channel and the pool
		close(logLineChan)
		pool.Close()
	}()
	// wait for all tasks to finish and close the pool
	for logLine := range logLineChan {
		// process the job
		logLine := logLine
		pool.Run(func() error {
			return f.ProcessLogLine(ctx, logLine)
		})
	}

	if err := pool.Wait(); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	return nil
}

func (f *fetchDataUC) readFile(ctx context.Context, bucket *storage.BucketHandle, file string, jobs chan<- []byte) error {
	// read the file
	// send the data to the jobs channel
	logctx.Infow(ctx, "reading file", "file", file)
	obj := bucket.Object(file).ReadCompressed(true)
	rdr, err := obj.NewReader(ctx)
	defer func() {
		if rdr != nil {
			rdr.Close()
		}
	}()
	if err != nil {
		logctx.Errorw(ctx, "failed to init object reader", err)
		return err
	}
	decompressed, err := gzip.NewReader(rdr)
	defer func() {
		logctx.Infow(ctx, "closing gzip reader", "file", file)
		if decompressed != nil {
			gzipCloseErr := decompressed.Close()
			if gzipCloseErr != nil {
				logctx.Errorw(ctx, "failed to closing gzip reader", "gzipCloseErr", gzipCloseErr)
			}
		}
	}()
	if err != nil {
		logctx.Errorw(ctx, "failed to init gzip reader", "err", err)
		return err
	}
	scanner := bufio.NewScanner(decompressed)
	buffer := bufferPool.Get().([]byte)
	scanner.Buffer(buffer, 5*1024*1024) // 5MB max line length
	defer bufferPool.Put(&buffer)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		line, err := f.preProcessLine(ctx, lineBytes)
		if err != nil {
			logctx.Errorw(ctx, "failed to pre process line", "err", err)
			continue
		}
		jobs <- line
		return nil // run only one job
	}
	return nil
}

func (f *fetchDataUC) preProcessLine(_ context.Context, line []byte) ([]byte, error) {
	// pre process the line
	skip, utf8Log := skipOrFixUtf8(line)
	if skip {
		return nil, errors.New("invalid utf8")
	}
	logLine := string(utf8Log)
	ll, err := strconv.Unquote(`"` + logLine + `"`)
	if err != nil {
		return nil, err
	}
	extractJSON, _, _ := f.extractJSON([]byte(ll))
	return extractJSON, nil
}

func (f *fetchDataUC) ProcessLogLine(ctx context.Context, logLine []byte) error {
	// process the job
	logctx.Infow(ctx, "processing job", "logLine", string(logLine))
	logObject, err := f.parseLogIntoStruct(ctx, string(logLine))
	if err != nil {
		return err
	}
	if err := f.storeLog(ctx, logObject); err != nil {
		return err
	}
	return nil
}

func skipOrFixUtf8(log []byte) (bool, []byte) {
	if utf8.ValidString(string(log)) {
		return false, log
	}
	utf8Reader := transform.NewReader(strings.NewReader(string(log)), unicode.UTF8.NewDecoder())
	utf8Log, err := io.ReadAll(utf8Reader)
	if err != nil {
		return true, nil
	}
	return false, utf8Log
}

func (f *fetchDataUC) parseLogIntoStruct(ctx context.Context, log string) (container.Map, error) {
	logObject := container.Map{}
	if err := json.Unmarshal([]byte(log), &logObject); err != nil {
		logctx.Errorw(ctx, "failed to unmarshal log", "err", err)
		return nil, err
	}
	if err := f.validateLogObject(logObject); err != nil {
		logctx.Errorw(ctx, "failed to validate log object", "err", err)
		return nil, err
	}
	return logObject, nil
}

/**
 * Extracts the JSON part from the log line.
 * remove the K8s prefix such as "2021-08-05T11:21:42.123456789Z stderr F ", or "2021-08-05T11:21:42.123456789Z stdout F "
 * the rest should be JSON otherwise return log line without k8s prefix
 */
func (f *fetchDataUC) extractJSON(logMsg []byte) (log []byte, k8sReceiveTime *time.Time, extractable bool) {
	logEntry := string(logMsg)

	// remove k8s prefix if present
	prefixes := k8sPrefixRegexp.FindStringSubmatch(logEntry)
	if len(prefixes) > 1 {
		logEntry = logEntry[len(prefixes[0]):]
		if len(prefixes) >= 2 {
			if t, err := time.Parse(time.RFC3339Nano, prefixes[1]); err == nil {
				k8sReceiveTime = &t
			}
		}
	}

	matches := jsonRegexp.FindStringSubmatch(logEntry)
	if len(matches) == 0 {
		return []byte(logEntry), k8sReceiveTime, false
	}

	// Return the JSON part
	return []byte(matches[0]), k8sReceiveTime, true
}
func (f *fetchDataUC) validateLogObject(logObject container.Map) error {
	if _, ok := logObject["request"]; !ok {
		return errors.New("missing request info")
	}
	if _, ok := logObject["request"].(map[string]any); !ok {
		return errors.New("request info is not a map")
	}
	if _, ok := logObject["response"]; !ok {
		return errors.New("missing response info")
	}
	if _, ok := logObject["response"].(map[string]any); !ok {
		return errors.New("response info is not a map")
	}
	return nil
}

func (f *fetchDataUC) storeLog(ctx context.Context, logObject container.Map) error {
	request := logObject["request"].(map[string]any)
	rawUrl := request["url"].(string)
	host, path := parseRawUrl(ctx, rawUrl)
	updatedPath := findParameterInPath(ctx, path)
	api, err := f.storage.GetApiByPath(ctx, updatedPath)
	if err != nil {
		return err
	}
	if api == nil {
		// create api
		api = &entity.Api{
			Host:                 host,
			Path:                 updatedPath,
			Method:               request["method"].(string),
			LatestBuildStructure: nil,
		}
		if err := f.storage.CreateApi(ctx, api); err != nil {
			return err
		}
	}
	if err := f.storeSampleRequest(ctx, api, request); err != nil {
		return err
	}
	response := logObject["response"].(map[string]any)
	if err := f.storeSampleResponse(ctx, api, response); err != nil {
		return err
	}
	return nil
}

func (f *fetchDataUC) storeSampleRequest(ctx context.Context, api *entity.Api, request container.Map) error {
	sampleRequest := &entity.SampleRequest{
		ApiId: api.Id,
		Body:  request["body"].(string),
	}
	if request["querystring"] != nil {
		sampleRequest.Parameters = buildParametersFromQueryString(request["querystring"].(map[string]any))
	}
	if err := f.storage.CreateSampleRequest(ctx, sampleRequest); err != nil {
		return err
	}
	return nil
}

func (f *fetchDataUC) storeSampleResponse(ctx context.Context, api *entity.Api, response container.Map) error {
	sampleResponse := &entity.SampleResponse{
		ApiId:          api.Id,
		HttpStatusCode: cast.ToInt(response["status"]),
		Body:           response["body"].(string),
	}
	if err := f.storage.CreateSampleResponse(ctx, sampleResponse); err != nil {
		return err
	}
	return nil
}
