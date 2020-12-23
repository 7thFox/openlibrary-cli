clean:
	rm -rf ./bin

build:
	go build -o ./bin/ol-cli ./*.go 

test: build bin/ol-cli
	cat test.txt | ./bin/ol-cli