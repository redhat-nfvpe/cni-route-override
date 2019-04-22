all: test testrace

deps:
	go get -d -v grpc.go4.org/...

updatedeps:
	go get -d -v -u -f grpc.go4.org/...

testdeps:
	go get -d -v -t grpc.go4.org/...

updatetestdeps:
	go get -d -v -t -u -f grpc.go4.org/...

build: deps
	go build grpc.go4.org/...

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go get -u -v github.com/golang/protobuf/protoc-gen-go
	# use $$dir as the root for all proto files in the same directory
	for dir in $$(git ls-files '*.proto' | xargs -n1 dirname | uniq); do \
		protoc -I $$dir --go_out=plugins=grpc:$$dir $$dir/*.proto; \
	done

test: testdeps
	go test -v -cpu 1,4 grpc.go4.org/...

testrace: testdeps
	go test -v -race -cpu 1,4 grpc.go4.org/...

clean:
	go clean -i grpc.go4.org/...

coverage: testdeps
	./coverage.sh --coveralls

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	proto \
	test \
	testrace \
	clean \
	coverage
