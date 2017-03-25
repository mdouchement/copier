package copier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/juju/errors"
)

type (
	// A Supervisor is batch copier handler.
	Supervisor struct {
		context.Context
		base          string
		logger        *Logger
		summary       map[string]int
		FilePaths     []string
		Destination   string
		ExecTimeout   time.Duration
		Progress      chan ProgressInfo
		Speed         float64
		Retries       int
		RetryInterval time.Duration
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
		Context: ctx,
		base:    base,
		logger:  NewLoggerWithContext(ctx),
		summary: map[string]int{
			StatusAlreadyExist: 0,
			StatusOverwritten:  0,
			StatusCopied:       0,
			StatusFailed:       0,
		},
		FilePaths:     paths,
		Destination:   destination,
		ExecTimeout:   10 * time.Minute,
		Progress:      make(chan ProgressInfo),
		Speed:         float64(512 * KBps),
		Retries:       5,
		RetryInterval: 2 * time.Second,
	}, err
}

// Logger returns the Supervisor logger.
func (s *Supervisor) Logger() *Logger {
	return s.logger
}

// Summary returns the summary of an execution.
func (s *Supervisor) Summary() map[string]int {
	return s.summary
}

// Execute runs the Supervisor copy batch.
func (s *Supervisor) Execute() error {
	fmt.Println("Start copy")
	fmt.Println("  - Speed:", int64(s.Speed), "Bps")
	fmt.Println("  - Timeout:", s.ExecTimeout)
	fmt.Println("  - Retries:", s.Retries)
	fmt.Println("  - Retry Interval:", s.RetryInterval)

	for _, from := range s.FilePaths {
		to := strings.Replace(from, s.base, s.Destination, 1)

		if err := s.execute(from, to, s.Retries, nil); err != nil {
			// Ignore returned error; go to next file.
			s.logger.C <- LogEntry{
				Severity:    ERROR,
				Status:      StatusFailed,
				FilepathSrc: from,
				FilepathDst: err.Error(),
			}
		}
	}
	return nil
}

func (s *Supervisor) execute(from, to string, retries int, err2 error) (err error) {
	if retries < 0 {
		return err2 // keep last error
	}

	errc := make(chan error)

	ctx, cancel := context.WithTimeout(s.Context, s.ExecTimeout)
	defer cancel()

	cp := NewExec(from, to, s.Speed)
	cp.Speed = s.Speed
	go func() {
		if err = cp.Execute(); err != nil {
			errc <- errors.Annotate(err, "Exec#Execute")
		}
		errc <- nil
	}()

	ready := false
	terminated := false
	for {
		select {
		case <-cp.Ready:
			s.Progress <- ProgressInfo{
				Name:        cp.Name(),
				Size:        cp.Size(),
				ProxyReader: cp.Reader(),
				Status:      cp.Status(),
			}
			ready = true
		case <-ctx.Done():
			if ctx.Err() != nil {
				cp.ForceClose()
				time.Sleep(s.RetryInterval)
				err = s.execute(from, to, retries-1, ctx.Err())
			}

			terminated = true
		case err = <-errc:
			if err != nil {
				cp.ForceClose()
				time.Sleep(s.RetryInterval)
				err = s.execute(from, to, retries-1, err)
			}

			terminated = true
		}

		if terminated {
			if !ready && err == nil {
				// Already exist
				s.Progress <- ProgressInfo{
					Name:        cp.Name(),
					Size:        cp.Size(),
					ProxyReader: cp.Reader(),
					Status:      cp.Status(),
				}
			}

			if err != nil {
				// failed
				s.Progress <- ProgressInfo{
					Name:        cp.Name(),
					Size:        cp.Size(),
					ProxyReader: cp.Reader(),
					Status:      StatusFailed,
				}
			}

			if err = s.logsOrReturnError(cp, err); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (s *Supervisor) logsOrReturnError(cp *Exec, err error) error {
	// Succeed
	if err == nil {
		s.summary[cp.Status()]++
		s.logger.C <- LogEntry{
			Severity:    INFO,
			Status:      cp.Status(),
			FilepathSrc: cp.From,
			FilepathDst: cp.To,
		}
		return nil
	}
	// Failed
	s.summary[StatusFailed]++
	if err == context.DeadlineExceeded || strings.HasSuffix(err.Error(), "would exceed context deadline") { // first is ctx.Err() and the second is a shapeio.Reder error
		s.logger.C <- LogEntry{
			Severity:    ERROR,
			Status:      StatusFailed,
			FilepathSrc: cp.From,
			FilepathDst: cp.To,
		}
		return nil
	}

	// Fatal error
	return err
}
