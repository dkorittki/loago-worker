package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

const (
	CacheDirName         = "loago_runner"
	networkEventChanSize = 300
)

type ChromeRunner struct {
	ID               uint
	CacheDir         string
	runFunc          func(ctx context.Context, actions ...chromedp.Action) error
	networkEventChan chan *network.EventResponseReceived
}

func NewChromeRunner(id uint) *ChromeRunner {
	r := &ChromeRunner{
		ID:               id,
		runFunc:          chromedp.Run,
		networkEventChan: make(chan *network.EventResponseReceived, networkEventChanSize),
	}

	return r
}

func (r *ChromeRunner) WithContext(ctx context.Context) context.Context {
	cachedir := filepath.Join(os.TempDir(), CacheDirName, fmt.Sprintf("%d", r.ID))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(cachedir),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	chromedpCtx, _ := chromedp.NewContext(allocCtx)

	r.CacheDir = cachedir
	runnerCtx := context.WithValue(chromedpCtx, contextKey{}, r)

	// Watch context and clean up browser cache once it's canceled
	f := cancel(runnerCtx)
	go f()

	if err := r.runFunc(runnerCtx, network.Enable()); err != nil {
		panic(err)
	}

	// Create a network event listener and send them into the runner buffer.
	// The Call() method will read and parse from it.
	chromedp.ListenTarget(chromedpCtx, func(ev interface{}) {
		if netEv, ok := ev.(*network.EventResponseReceived); ok {
			if netEv.Type == network.ResourceTypeDocument {
				r.networkEventChan <- netEv
			}
		}
	})

	return runnerCtx
}

func cancel(ctx context.Context) func() {
	return func() {
		v := FromContext(ctx)
		r := v.(*ChromeRunner)

		select {
		case <-ctx.Done():
			log.Debug().
				Str("component", "runner").
				Uint("id", r.ID).
				Str("cachedir", r.CacheDir).
				Msg("delete cache")

			var err error
			for i := 0; i < 10; i++ {
				err = os.RemoveAll(r.CacheDir)
				if err == nil {
					return
				}
				time.Sleep(200 * time.Millisecond)
			}

			log.Warn().
				Str("component", "runner").
				Uint("id", r.ID).
				Err(err).
				Msg("can't delete cache")

			// close network event buffer.
			close(r.networkEventChan)
		}
	}
}