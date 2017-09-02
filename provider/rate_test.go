package provider

import "testing"

func TestReqRateObserve(t *testing.T) {
	t.Parallel()

	const offset = 128

	rate := &ReqRate{offset: offset}
	rate.observe(offset + (blockSize - nquants) + 1)
	rate.observe(offset + (blockSize - nquants) + 1)
	rate.observe(offset + 2*blockSize)
	rate.observe(offset + blockSize)
	rate.observe(offset + 2*blockSize)

	if nreq := rate.rate(offset + blockSize); nreq != 3 {
		t.Fatalf("Invalid rate expected : 3, but actual : %v", nreq)
	}
	if nreq := rate.rate(offset + 2*blockSize); nreq != 2 {
		t.Fatalf("Invalid rate expected : 2, but actual : %v", nreq)
	}
}

func TestReqRateClean(t *testing.T) {
	t.Parallel()

	const offset = 128

	rate := &ReqRate{offset: offset}
	rate.observe(offset)
	rate.observe(offset + blockSize)
	rate.observe(offset + 2*blockSize)

	rate.clean(offset + 3*blockTTL)

	if rate.head != nil {
		t.Fatal("Clean doesn't delete all expired blocks")
	}
}
