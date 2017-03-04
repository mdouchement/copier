package copier

import (
	"context"
	"io"
	"os"

	"github.com/fujiwara/shapeio"
	"github.com/juju/errors"
)

// An Exec copies files at a given ratelimit.
type Exec struct {
	ctx    context.Context
	reader *ProxyReader
	From   string
	To     string
	Speed  float64
	status string
	Ready  chan struct{}
}

// NewExec returns a new Exec.
func NewExec(from, to string, speed float64) *Exec {
	return NewExecWithContext(context.Background(), from, to, speed)
}

// NewExecWithContext returns a new Exec with the given ctx.
func NewExecWithContext(ctx context.Context, from, to string, speed float64) *Exec {
	return &Exec{
		ctx:   ctx,
		From:  from,
		To:    to,
		Speed: speed,
		Ready: make(chan struct{}),
	}
}

// Execute starts the copy with a ratelimit.
func (e *Exec) Execute() error {
	MkdirAllWithFilename(e.To)

	ok, err := MoreRecent(e.To, e.From)
	if err != nil {
		e.Ready <- struct{}{}
		return err
	}

	if ok {
		e.status = StatusAlreadyExist
		e.Ready <- struct{}{}
		return nil
	}
	if Exists(e.To) {
		e.status = StatusOverwritten
	} else {
		e.status = StatusCopied
	}

	r, err := os.Open(e.From)
	if err != nil {
		e.Ready <- struct{}{}
		return errors.Annotate(err, "source")
	}
	defer r.Close()

	w, err := os.Create(e.To)
	if err != nil {
		e.Ready <- struct{}{}
		return errors.Annotate(err, "destination")
	}
	defer w.Close()

	rr := shapeio.NewReaderWithContext(r, e.ctx)
	rr.SetRateLimit(e.Speed)

	e.reader = NewProxyReader(rr) // Used for progressbar
	defer e.reader.Close()
	e.Ready <- struct{}{}

	if _, err = io.Copy(w, e.reader); err != nil {
		return errors.Annotate(err, "copy")
	}

	return w.Sync()
}

// Status returns the copy status.
func (e *Exec) Status() string {
	return e.status
}

// Name returns the filename.
func (e *Exec) Name() string {
	return e.From
}

// Size returns the file size.
func (e *Exec) Size() int64 {
	return Size(e.From)
}

// Reader returns the file reader.
func (e *Exec) Reader() *ProxyReader {
	return e.reader
}
