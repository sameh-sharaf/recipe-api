default: build

deps:
	go get -u github.com/gorilla/mux
	go get -u github.com/lib/pq
	go get -u golang.org/x/crypto/bcrypt

bin/api-test: src/*.go
	env GOOS=linux GOARCH=386 go build -o bin/api-test $^

build: bin/api-test

run: build
	bin/api-test

clean:
	rm -f bin/api-test
	rmdir bin
