package tokenbucket

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity  int64
	tokens    chan struct{}
	rate      time.Duration // Add a token to the bucket every 1/r units of time
	rateMutex sync.Mutex
}

func NewBucket(rate time.Duration, capacity int64) *Bucket {

	//A bucket is simply a channel with a buffer representing the maximum size
	tokens := make(chan struct{}, capacity)

	b := &Bucket{capacity, tokens, rate, sync.Mutex{}}

	//Set off a function that will continuously add tokens to the bucket
	go func(b *Bucket) {
		ticker := time.NewTicker(rate)
		for _ = range ticker.C {
			b.tokens <- struct{}{}
		}
	}(b)

	return b
}

func (b *Bucket) GetRate() time.Duration {
	b.rateMutex.Lock()
	tmp := b.rate
	b.rateMutex.Unlock()
	return tmp
}

func (b *Bucket) SetRate(rate time.Duration) {
	b.rateMutex.Lock()
	b.rate = rate
	b.rateMutex.Unlock()
}

//AddTokens manually adds n tokens to the bucket
func (b *Bucket) AddToken(n int64) {
}

func (b *Bucket) withdrawTokens(n int64) error {
	for i := int64(0); i < n; i++ {
		<-b.tokens
	}
	return nil
}

func (b *Bucket) SpendToken(n int64) <-chan error {
	// Default to spending a single token
	if n < 0 {
		n = 1
	}

	c := make(chan error)
	go func(b *Bucket, n int64, c chan error) {
		c <- b.withdrawTokens(n)
		close(c)
		return
	}(b, n, c)

	return c
}

// Drain will empty all tokens in the bucket
// If the tokens are being added too quickly (if the rate is too fast)
// this will never drain
func (b *Bucket) Drain() error{
    // TODO replace this with a more solid approach (such as replacing the channel altogether)
    for {
        select {
            case _ = <-b.tokens:
                continue
            default:
                return nil
        }
    }
}
