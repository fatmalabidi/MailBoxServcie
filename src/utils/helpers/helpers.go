package helpers

import (
	"context"
	"os"
	"time"

	"github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/sirupsen/logrus"
)

const (
	USERNAME string = "AUTH_USERNAME"
	PASSWORD string = "AUTH_PASSWORD"
)

// IdxRange defines lower an upper bound of slice
type IdxRange struct {
	Low, High int
}

// Partition slices a slice and returns lower an upper bound of each sub slice
func Partition(collectionLen, partitionSize int) chan IdxRange {
	c := make(chan IdxRange)
	if partitionSize <= 0 {
		close(c)
		return c
	}

	go func() {
		numFullPartitions := collectionLen / partitionSize
		i := 0
		for ; i < numFullPartitions; i++ {
			c <- IdxRange{Low: i * partitionSize, High: (i + 1) * partitionSize}
		}

		if collectionLen%partitionSize != 0 { // left over
			c <- IdxRange{Low: i * partitionSize, High: collectionLen}
		}
		close(c)
	}()
	return c
}

// GetLogger configures the logging and return a logger
func GetLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:       false,
		DisableHTMLEscape: true,
		TimestampFormat:   time.RFC3339,
	})
	logger.SetReportCaller(true)
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

// GetTestLogger configures the testing logging and return a logger
func GetTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:       false,
		DisableHTMLEscape: true,
		TimestampFormat:   time.RFC3339,
	})
	logger.SetReportCaller(true)
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

// AddTimeoutToCtx add timeout to context.
func AddTimeoutToCtx(ctx context.Context, timeOutDuration int64) (context.Context, context.CancelFunc) {
	timeOut := time.Duration(timeOutDuration) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	return ctx, cancel
}

// EnrichLog function receives a logger and returned it after adding to it fields from context and from extra map
func EnrichLog(l *logrus.Logger, ctx context.Context, extra map[string]string) *logrus.Entry {
	var logger *logrus.Entry
	// add logs from context

	cid, ok := ctx.Value(models.CORRELATION_ID).(string)
	if ok {
		logger = l.WithField(models.CORRELATION_ID, cid)
	}

	// add logs from extra
	for key, value := range extra {
		logger = l.WithField(key, value)
	}
	return logger
}

func GetCredFromEnv() (string, string) {
	username, _ := os.LookupEnv(USERNAME)
	password, _ := os.LookupEnv(PASSWORD)
	return username, password
}

// SetCredToEnv function injects the received credentials to the environment
func SetCredToEnv(username string, password string) error {
	err := os.Setenv(USERNAME, username)
	if err != nil {
		return err
	}
	err = os.Setenv(PASSWORD, password)
	if err != nil {
		return err
	}
	return nil
}
