test:
	go test ./... -count=1 -cover

test-integration:
	go test ./... -count=1 --tags=local,integration -cover

run:
	go run .

install_golint:
	go get -u golang.org/x/lint/golint

lint:
	@echo Ensure you have the golint official CLI or this command will fail.
	@echo You can install the golint CLI with: make install_golint
	@echo ....

	golint ./...