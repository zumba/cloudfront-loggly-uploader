package cloudfront

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/s3"
	"github.com/zumba/cloudfront-loggly-uploader/logging"
	"github.com/zumba/cloudfront-loggly-uploader/loggly"
)

// LogEntry is a parsed CloudFront log line
type LogEntry struct {
	Date            string `json:"date"`
	Time            string `json:"time"`
	Timestamp       string `json:"timestamp"`
	XEdgeLocation   string `json:"x-edge-location"`
	ScBytes         int    `json:"sc-bytes"`
	ClientIP        string `json:"client-ip"`
	Method          string `json:"method"`
	CsHost          string `json:"cs-host"`
	Request         string `json:"request"`
	Status          string `json:"status"`
	Referer         string `json:"referer"`
	UserAgent       string `json:"user-agent"`
	CsURIQuery      string `json:"cs-uri-query"`
	CsCookie        string `json:"cs-cookie"`
	XEdgeResultType string `json:"x-edge-result-type"`
	XEdgeRequestID  string `json:"x-edge-request-id"`
	XHostHeader     string `json:"x-host-header"`
	CsProtocol      string `json:"cs-protocol"`
	CsBytes         string `json:"cs-bytes"`
	TimeTaken       string `json:"time-taken"`
	XForwardedFor   string `json:"x-forwarded-for"`
	SslProtocol     string `json:"ssl-protocol"`
	SslCipher       string `json:"ssl-cipher"`
	Cache           string `json:"cache"`
}

// Processor struct contains variables used to process CloudFront logs
type Processor struct {
	s3Service    *s3.S3
	s3Bucket     string
	s3Prefix     string
	logglyClient *loggly.Client
}

// NewLogEntry creates new LogEntry struct by parsing space separated line
func NewLogEntry(line string) (*LogEntry, error) {
	if len(line) == 0 {
		return nil, errors.New("Can't create LogEntry from empty line")
	}

	splitRegexp := regexp.MustCompile("\\s+")
	splitResult := splitRegexp.Split(line, -1)
	splitResultFieldCount := len(splitResult)

	cfLogEntry := &LogEntry{}

	cfLogEntry.Date = splitResult[0]
	cfLogEntry.Time = splitResult[1]
	cfLogEntry.Timestamp = fmt.Sprintf("%sT%sZ", cfLogEntry.Date, cfLogEntry.Time)
	cfLogEntry.XEdgeLocation = splitResult[2]

	// Try converting scBytes to integer
	scBytesInt, err := strconv.Atoi(splitResult[3])
	if err != nil {
		cfLogEntry.ScBytes = scBytesInt
	} else {
		cfLogEntry.ScBytes = 0
	}

	cfLogEntry.ClientIP = splitResult[4]
	cfLogEntry.Method = splitResult[5]
	cfLogEntry.CsHost = splitResult[6]
	cfLogEntry.Request = splitResult[7]
	cfLogEntry.Status = splitResult[8]
	cfLogEntry.Referer = splitResult[9]
	cfLogEntry.UserAgent = splitResult[10]
	cfLogEntry.CsURIQuery = splitResult[11]
	cfLogEntry.CsCookie = splitResult[12]
	cfLogEntry.XEdgeResultType = splitResult[13]
	cfLogEntry.XEdgeRequestID = splitResult[14]
	cfLogEntry.XHostHeader = splitResult[15]
	cfLogEntry.CsProtocol = splitResult[16]
	cfLogEntry.CsBytes = splitResult[17]
	cfLogEntry.TimeTaken = splitResult[18]
	// Use the number of fields counted to determine which CloudFront format
	// we're dealing with. One format has 20 fields, the other has 23.
	if splitResultFieldCount > 19 {
		cfLogEntry.XForwardedFor = splitResult[19]
		cfLogEntry.SslProtocol = splitResult[20]
		cfLogEntry.SslCipher = splitResult[21]
		cfLogEntry.Cache = splitResult[22]
	}

	return cfLogEntry, nil
}

// JSON Marshals the LogEntry struct into JSON and returns the resulting string
func (le *LogEntry) JSON() (string, error) {
	jsonBytes, err := json.Marshal(le)
	if err != nil {
		return "", fmt.Errorf("Unable to marshal to JSON: %v", le)
	}

	jsonString := string(jsonBytes) + "\n"

	return jsonString, nil
}

// NewProcessor creates a new Processor object
func NewProcessor(s3Service *s3.S3, s3Bucket string, s3Prefix string, logglyClient *loggly.Client, debug bool) *Processor {
	processor := Processor{}
	processor.s3Service = s3Service
	processor.s3Bucket = s3Bucket
	processor.s3Prefix = s3Prefix
	processor.logglyClient = logglyClient

	return &processor
}

// GetListOfS3FilesInBucket gets a list of S3 files in a bucket with prefix
func (p *Processor) GetListOfS3FilesInBucket() []string {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(p.s3Bucket),
		Prefix: aws.String(p.s3Prefix),
	}

	resp, err := p.s3Service.ListObjects(params)

	if err != nil {
		log.Println(err)
	}

	keyList := make([]string, len(resp.Contents))
	for i := range resp.Contents {
		keyList[i] = *resp.Contents[i].Key
	}

	return keyList
}

// GetS3FileContent gets the contents of a specific file/key. Most likely
// gzip compressed.
func (p *Processor) GetS3FileContent(key string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: &p.s3Bucket,
		Key:    &key,
	}

	resp, err := p.s3Service.GetObject(params)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var bytes bytes.Buffer
	_, err = io.Copy(&bytes, resp.Body)

	return bytes.Bytes(), nil
}

// DeleteS3File delete a specific S3 file
func (p *Processor) DeleteS3File(key string) error {
	params := &s3.DeleteObjectInput{
		Bucket: &p.s3Bucket,
		Key:    &key,
	}

	_, err := p.s3Service.DeleteObject(params)

	if err != nil {
		return err
	}

	return nil
}

// UncompressS3File uncompresses the S3 file content and returns the string value
func UncompressS3File(compressedBytes []byte) (string, error) {
	compressedBuffer := bytes.NewBuffer(compressedBytes)

	gzipReader, err := gzip.NewReader(compressedBuffer)
	defer gzipReader.Close()

	var uncompressedBuffer bytes.Buffer
	_, err = io.Copy(&uncompressedBuffer, gzipReader)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(uncompressedBuffer.Bytes()), nil
}

// ProcessLogsInBucket processes the CloudFront logs found in bucket
func (p *Processor) ProcessLogsInBucket() {
	s3FileList := p.GetListOfS3FilesInBucket()

	log.Printf("Number of S3 files to process: %d\n", len(s3FileList))
	log.Printf("List of files: %v\n", s3FileList)

	for _, s3File := range s3FileList {

		log.Printf("Processing S3 file: %s", s3File)

		compressedContent, err := p.GetS3FileContent(s3File)
		if err != nil {
			log.Printf("Unable to get S3 file content for file %s %v", s3File, err)
		}

		uncompressedContent, err := UncompressS3File(compressedContent)
		if err != nil {
			log.Printf("Unable to uncompress %s %v", s3File, err)
		}

		err = p.processCloudFrontLogFile(uncompressedContent)

		if err != nil {
			log.Printf("Unable to process CloudFront log file: %s", s3File)
		}

		log.Printf("Deleting CloudFront S3 log file: %s", s3File)

		err = p.DeleteS3File(s3File)

		if err != nil {
			log.Printf("Unable to delete CloudFront S3 file after processing it: %s", err)
		}
	}
}

// processCloudFrontLogFile parses an uncompressed CloudFront log file
func (p *Processor) processCloudFrontLogFile(input string) error {

	for _, v := range strings.Split(input, "\n") {
		logging.DebugLogger.Printf("Processing line: %s\n", v)

		// skip empty lines
		if len(v) == 0 {
			logging.DebugLogger.Println("Skipping empty line")
			continue
		}

		// skip comments
		if v[0] == '#' {
			logging.DebugLogger.Println("Skipping comment line")
			continue
		}

		cfLog, err := NewLogEntry(v)
		if err != nil {
			log.Printf("Unable to parse line: %v", v)
			continue
		}

		logging.DebugLogger.Printf("Parsed CloudFrontLine = %v\n", cfLog)

		jsonString, err := cfLog.JSON()
		if err != nil {
			log.Printf("Unable to Marshal cfLog to JSON\n")
			continue
		}

		p.logglyClient.AddToBuffer(jsonString)
	}
	// Send anything left in buffer
	p.logglyClient.SendBuffer()

	return nil
}
