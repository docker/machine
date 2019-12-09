package amazonec2

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
)

type awslogger struct {
	logger *log.Logger
}

func AwsLogger() aws.Logger {
	return &awslogger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (l awslogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}
