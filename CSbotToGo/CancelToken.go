package CSbotToGo

import (
	"context"
	"time"
)

type CancellationToken struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewCancellationToken() *CancellationToken {
	ctx, cancel := context.WithCancel(context.Background())
	return &CancellationToken{ctx: ctx, cancel: cancel}
}

func (ct *CancellationToken) IsCancellationRequested() bool {
	select {
	case <-ct.ctx.Done():
		return true
	default:
		return false
	}
}

func (ct *CancellationToken) Token() context.Context {
	return ct.ctx
}

func (ct *CancellationToken) Cancel() {
	ct.cancel()
}

type CancellationTokenSource struct {
	ct *CancellationToken
}

func NewCancellationTokenSource() *CancellationTokenSource {
	return &CancellationTokenSource{ct: NewCancellationToken()}
}

func (cts *CancellationTokenSource) Token() *CancellationToken {
	return cts.ct
}

func (cts *CancellationTokenSource) Cancel() {
	cts.ct.Cancel()
}

func (cts *CancellationTokenSource) CancelAfter(duration time.Duration) {
	timer := time.NewTimer(duration)
	go func() {
		<-timer.C
		cts.Cancel()
	}()
}
