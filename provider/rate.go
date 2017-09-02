package provider

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	blockSize = 64
	quant     = time.Second
	interval  = time.Minute

	nquants = int64(interval / quant)

	blockTTL = 4 * blockSize // in seconds
)

var (
	rateBlockPool = sync.Pool{
		New: func() interface{} { return new(RateBlock) },
	}
)

// RateBlock ...
type RateBlock struct {
	counts [blockSize]int64 // i-cell is number of requests for the (offset+i)-th second
	offset int64            // time offset in Unix seconds
	next   unsafe.Pointer   // real type is *RateBlock
}

func loadBlock(ptr *unsafe.Pointer) *RateBlock {
	return (*RateBlock)(atomic.LoadPointer(ptr))
}

func blockFromPool() *RateBlock {
	block := rateBlockPool.Get().(*RateBlock)

	counts := block.counts[:]
	for i := range counts {
		counts[i] = 0
	}
	block.offset, block.next = 0, nil

	return block
}

// Next ...
func (b *RateBlock) Next() *RateBlock {
	return loadBlock(&b.next)
}

// ReqRate - lock-free datastructure for counting number of requests for the last minute
type ReqRate struct {
	head   unsafe.Pointer // real type is *RateBlock
	offset int64          // global offset
}

// NewReqRate - constructor for ReqRate struct
func NewReqRate() *ReqRate {
	return &ReqRate{
		offset: time.Now().Unix(),
	}
}

// Head ...
func (r *ReqRate) Head() *RateBlock {
	return loadBlock(&r.head)
}

func (r *ReqRate) clean(now int64) {
	deadline := now - blockTTL

	for head := r.Head(); head != nil && head.offset < deadline; head = r.Head() {
		next := atomic.LoadPointer(&head.next)
		if atomic.CompareAndSwapPointer(&r.head, unsafe.Pointer(head), next) {
			rateBlockPool.Put(head)
		}
	}
}

func (r *ReqRate) observe(now int64) {
	// help clean up ReqRate struct
	r.clean(now)

	indirect := &r.head
	for {
		this := loadBlock(indirect)

		if this == nil || this.offset > now {
			block := blockFromPool()
			block.offset = r.offset + (((now - r.offset) / nquants) * nquants)
			block.next = unsafe.Pointer(this)

			// register request in the new block
			index := now - block.offset
			block.counts[index] = 1

			// try add new block to block list
			if atomic.CompareAndSwapPointer(indirect, block.next, unsafe.Pointer(block)) {
				return
			}

			// failed to add block (maybe needed block has been added)
			rateBlockPool.Put(block)
			continue
		}

		// this block is needed, so increment count
		if index := now - this.offset; 0 <= index && index < blockSize {
			atomic.AddInt64(&this.counts[index], 1)
			return
		}

		indirect = &this.next
	}

}

// Observe - register request with this time
func (r *ReqRate) Observe(now time.Time) {
	r.observe(time.Now().Unix())
}

func (r *ReqRate) rate(now int64) (sum int64) {
	// help clean up ReqRate struct
	r.clean(now)

	// quanted time interval [since, until]
	since, until := (now - nquants), now

	indirect := &r.head
	for this := loadBlock(indirect); this != nil && this.offset <= until; this = loadBlock(indirect) {
		var left, right int64
		if index := since - this.offset; 0 <= index && index < blockSize {
			left = index
		}
		if index := until - this.offset; 0 <= index && index < blockSize {
			right = index
		}

		counts := this.counts[left:right]
		for i := range counts {
			count := atomic.LoadInt64(&counts[i])
			sum += count
		}

		indirect = &this.next
	}
	return
}

// Rate returns request number for the last minute
func (r *ReqRate) Rate() int64 {
	return r.rate(time.Now().Unix())
}
