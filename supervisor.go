package copier

import (
	"context"
	"strings"
	"time"

	"github.com/juju/errors"
)

type (
	// A Supervisor is batch copier handler.
	Supervisor struct {
		context.Context
		base        string
		logger      *Logger
		FilePaths   []string
		Destination string
		ExecTimeout time.Duration
		Progress    chan ProgressInfo
		Speed       float64
	}

	// ProgressInfo handle data for building a progress bar.
	ProgressInfo struct {
		Name        string
		Size        int64
		ProxyReader *ProxyReader
		Status      string
	}
)

// NewSupervisor returns a new Supervisor.
func NewSupervisor(paths []string, destination string) (*Supervisor, error) {
	base, err := GetBaseDirectory(paths)
	ctx := context.Background()
	return &Supervisor{
		Context:     ctx,
		base:        base,
		logger:      NewLoggerWithContext(ctx),
		FilePaths:   paths,
		Destination: destination,
		ExecTimeout: 10 * time.Minute,
		Progress:    make(chan ProgressInfo),
		Speed:       float64(512 * KBps),
	}, err
}

// Execute runs the Supervisor copy batch.
func (s *Supervisor) Execute() error {
	errc := make(chan error)

	for _, from := range s.FilePaths {
		to := strings.Replace(from, s.base, s.Destination, 1)

		ctx, cancel := context.WithTimeout(context.Background(), s.ExecTimeout)
		defer cancel() // FIXME

		cp := NewExecWithContext(ctx, from, to, s.Speed)
		cp.Speed = s.Speed
		go func() {
			if err := cp.Execute(); err != nil {
				errc <- errors.Annotate(err, "Exec#Execute")
			}
			errc <- nil
		}()

		<-cp.Ready

		s.Progress <- ProgressInfo{
			Name:        cp.Name(),
			Size:        cp.Size(),
			ProxyReader: cp.Reader(),
			Status:      cp.Status(),
		}

		if err := <-errc; err != nil {
			if strings.HasSuffix(err.Error(), "would exceed context deadline") {
				s.logger.C <- LogEntry{
					Severity:    ERROR,
					Status:      StatusFailed,
					FilepathSrc: cp.From,
					FilepathDst: cp.To,
				}
			} else {
				return err
			}
		} else {
			s.logger.C <- LogEntry{
				Severity:    INFO,
				Status:      cp.Status(),
				FilepathSrc: cp.From,
				FilepathDst: cp.To,
			}
		}
	}
	return nil
}

// Logger returns the Supervisor logger.
func (s *Supervisor) Logger() *Logger {
	return s.logger
}
