make: syntax.peg.go yasp

test: make
	go test

yasp:
	go build

syntax.peg.go: syntax.peg
	~/bin/gobin/peg -switch -inline syntax.peg

clean:
	rm -f yasp syntax.peg.go
