package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mdouchement/copier"
)

type formatter struct{}

func (f *formatter) Format(entry *log.Entry) ([]byte, error) {
	fields := []string{}
	for k, v := range entry.Data {
		fields = append(fields, fmt.Sprintf("%s=%s", k, v))
	}

	data := fmt.Sprintf("[%s] %+5s: %s (%s)\n",
		time.Now().Format(time.RFC1123),
		strings.ToUpper(entry.Level.String()),
		entry.Message,
		strings.Join(fields, ", "),
	)
	return []byte(data), nil
}

// StartLogger starts the copier CLI logger.
func StartLogger(l *copier.Logger, filename string) error {
	log.SetFormatter(new(formatter))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	log.SetOutput(file)

	go func() {
		for {
			select {
			case <-l.Done():
				file.Close()
				return
			case entry := <-l.C:
				msg := fmt.Sprintf("%s  ->  %s", entry.FilepathSrc, entry.FilepathDst)
				switch entry.Severity {
				case copier.DEBUG:
					log.WithField("status", entry.Status).Debug(msg)
				case copier.INFO:
					log.WithField("status", entry.Status).Info(msg)
				case copier.WARN:
					log.WithField("status", entry.Status).Warn(msg)
				case copier.ERROR:
					log.WithField("status", entry.Status).Error(msg)
				}
			}
		}
	}()

	return nil
}
