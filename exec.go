package copier

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fujiwara/shapeio"
	"github.com/juju/errors"
)

// An Exec copies files at a given ratelimit.
type Exec struct {
	opened bool
	size   int64
	ctx    context.Context
	r      io.ReadCloser  // input file
	w      io.WriteCloser // output file
	pr     *ProxyReader
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
	s, err := Size(e.From)
	if err != nil {
		return errors.Annotate(err, "source")
	}
	e.size = s

	MkdirAllWithFilename(e.To)

	ok, err := MoreRecent(e.To, e.From)
	if err != nil {
		return err
	}

	if ok {
		e.status = StatusAlreadyExist
		return nil
	}
	if Exists(e.To) {
		e.status = StatusOverwritten
	} else {
		e.status = StatusCopied
	}

	e.opened = true

	e.r, err = os.Open(e.From)
	if err != nil {
		return errors.Annotate(err, "source")
	}
	defer e.r.Close()

	w, err := os.Create(e.To)
	if err != nil {
		return errors.Annotate(err, "destination")
	}
	defer w.Close()
	e.w = w

	rr := shapeio.NewReaderWithContext(e.r, e.ctx)
	rr.SetRateLimit(e.Speed)

	e.pr = NewProxyReader(rr) // Used for progressbar
	defer e.pr.Close()
	e.Ready <- struct{}{}

	if _, err = io.Copy(e.w, e.pr); err != nil {
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
	return e.size
}

// Reader returns the file reader.
func (e *Exec) Reader() *ProxyReader {
	return e.pr
}

// ForceClose force closes all its IO objects.
func (e *Exec) ForceClose() {
	if !e.opened {
		return
	}

	e.asyncClose(e.pr)
	e.asyncClose(e.r)
	e.asyncClose(e.w)
	fmt.Println("Forced close.")
}

func (e *Exec) asyncClose(c io.Closer) {
	if c != nil {
		go c.Close()
	}
}
