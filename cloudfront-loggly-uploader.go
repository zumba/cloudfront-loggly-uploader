// Main package
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"strconv"
	"time"

	"github.com/zumba/cloudfront-loggly-uploader/cloudfront"
	"github.com/zumba/cloudfront-loggly-uploader/logging"
	"github.com/zumba/cloudfront-loggly-uploader/loggly"

	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/session"
	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/s3"
	"github.com/zumba/cloudfront-loggly-uploader/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// Config contains config variables for application
type Config struct {
	DaemonSleepMins      string `yaml:"daemon-sleep-mins"`
	S3Region             string `yaml:"s3-region"`
	S3Bucket             string `yaml:"s3-bucket"`
	S3Prefix             string `yaml:"s3-prefix"`
	S3AWSAccessKeyID     string `yaml:"s3-aws-access-key-id"`
	S3AWSSecretAccessKey string `yaml:"s3-aws-secret-access-key"`
	LogglyAPIKey         string `yaml:"loggly-api-key"`
	LogglyTag            string `yaml:"loggly-tag"`
}

var configFile = flag.String("configFile", "/etc/cloudfront-loggly-uploader.conf.yaml", "Config file")
var syslogFlag = flag.Bool("syslog", false, "Log to syslog instead of stdout")
var daemonFlag = flag.Bool("daemon", false, "Run in daemon mode")
var debugFlag = flag.Bool("debug", false, "Enable debug output")

func setupLogging() {
	if *syslogFlag == true {
		logToSyslog()
	} else {
		logToStdout()
	}
}

func logToStdout() {
	log.SetOutput(os.Stdout)
	if *debugFlag == true {
		log.SetFlags(log.Flags() | log.Lshortfile)
		logging.DebugLogger.SetFlags(log.Lshortfile)
		logging.DebugLogger.SetOutput(os.Stdout)
	}
}

func logToSyslog() {
	syslogWriter, err := syslog.New(syslog.LOG_NOTICE, os.Args[0])
	if err != nil {
		log.Fatalf("Unable to create syslog writer: %s\n", err)
	}

	log.SetFlags(0)
	log.SetOutput(syslogWriter)
	if *debugFlag == true {
		logging.DebugLogger.SetFlags(log.Lshortfile)
		logging.DebugLogger.SetOutput(syslogWriter)
	}
}

// Main function
func main() {
	flag.Parse()

	setupLogging()

	appConfig := &Config{}

	configFileData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Printf("Unable to open config file: %s\n", *configFile)
		os.Exit(1)
	}

	err = yaml.Unmarshal(configFileData, &appConfig)
	if err != nil {
		log.Printf("Unable to unmarshal Yaml config file: %s\n", *configFile)
		os.Exit(1)
	}

	awsCreds := credentials.NewStaticCredentials(appConfig.S3AWSAccessKeyID,
		appConfig.S3AWSSecretAccessKey,
		"")

	awsConfig := aws.NewConfig().WithCredentials(awsCreds).WithRegion(appConfig.S3Region)

	s3Service := s3.New(session.New(), awsConfig)

	logglyClient := loggly.NewClient(appConfig.LogglyAPIKey,
		appConfig.LogglyTag,
		*debugFlag)

	cloudFrontProcessor := cloudfront.NewProcessor(s3Service,
		appConfig.S3Bucket,
		appConfig.S3Prefix,
		logglyClient,
		*debugFlag)

	if *daemonFlag == true {
		sleepMins, err := strconv.Atoi(appConfig.DaemonSleepMins)
		if err != nil {
			log.Fatalf("Unable to convert DaemonSleepMins to integer: %s\n", err)
		}

		log.Printf("Running in daemon mode. Checking for new files every %d mins\n", sleepMins)

		// Run indefinitely
		for {
			cloudFrontProcessor.ProcessLogsInBucket()
			sleepDuration := time.Minute * time.Duration(sleepMins)
			log.Printf("Finished processing CloudFront logs. Sleeping: %s\n", sleepDuration)
			time.Sleep(sleepDuration)
		}
	} else {
		cloudFrontProcessor.ProcessLogsInBucket()
	}
}
