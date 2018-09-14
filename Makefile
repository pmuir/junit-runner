build: clean
	go build -o bin/junit-runner-osx
	GOOS=linux GOARCH=386 go build -o bin/junit-runner-linux

clean:
	rm -rf bin/
