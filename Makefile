make: syntax.peg.go yasp

get-deps:
	go get github.com/pointlander/peg

test: make
	go test ./tests/assertions.go ./tests/main_test.go

yasp:
	go build ./cmd/yasp

syntax.peg.go: src/syntax.peg
	${GOPATH}/bin/peg -switch -inline src/syntax.peg

clean:
	rm -f yasp src/syntax.peg.go
