make: syntax.peg.go yasp

get-deps:
	go get github.com/pointlander/peg

test: make
	go test ./tests/...

yasp:
	go build

syntax.peg.go: src/syntax.peg
	${GOPATH}/bin/peg -switch -inline src/syntax.peg

clean:
	rm -f yasp src/syntax.peg.go
