.PHONY: tests
tests:
	go test -cover -v searchinform/cache \
				searchinform/provider
