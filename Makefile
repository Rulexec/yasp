make: syntax.peg.go yasp

get-deps:
	go get github.com/pointlander/peg

test: make
	go test ./tests/...

yasp:
	go build

syntax.peg.go: syntax.peg
	${GOPATH}/bin/peg -switch -inline syntax.peg

clean:
	rm -f yasp syntax.peg.go
