build:
	mkdir -p bin
	go build -o bin/issuer -v ./cmd/issuer
	go build -o bin/acquirer -v ./cmd/acquirer

check:
	go test ./...

.PHONY: acquirer
acquirer:
	go run cmd/acquirer/main.go

.PHONY: issuer
issuer:
	go run cmd/issuer/main.go

.PHONY: card-personalizer
card-personalizer:
	JAVA_HOME=/opt/homebrew/opt/openjdk@11 go run ./cmd/cardpersonalizer
	# go run cmd/cardpersonalizer/main.go

.PHONY: printer
printer:
	go run cmd/printer/main.go

.PHONY: terminal
terminal:
	go run cmd/terminal/main.go


.PHONY: presenter-acquirer
presenter-acquirer:
	@echo "Starting ngrok and acquirer..."
	@trap 'kill 0' INT; \
	ngrok http --domain=ftdc-acquirer.ngrok.io --log=stdout 8080 2>&1 & \
	go run cmd/acquirer/main.go 2>&1 & \
	wait

.PHONY: presenter-issuer
presenter-issuer:
	@echo "Starting ngrok and issuer..."
	@trap 'kill 0' INT; \
	ngrok http --domain=ftdc-issuer.ngrok.io --log=stdout 2>&1 9090 2>&1 & \
	ngrok tcp --region=us --remote-addr=5.tcp.ngrok.io:27433 --log=stdout 2>&1 8583 2>&1 & \
	go run cmd/issuer/main.go 2>&1 & \
	wait

.PHONY: presenter-card-personalizer
presenter-card-personalizer:
	@echo "Starting ngrok and issuer..."
	@trap 'kill 0' INT; \
	ngrok http --domain=ftdc-card-maker.ngrok.io --log=stdout 2>&1 7070 2>&1 & \
	JAVA_HOME=/opt/homebrew/opt/openjdk@11 go run cmd/cardpersonalizer/main.go 2>&1 & \
	wait

.PHONY: presenter-printer
presenter-printer:
	@echo "Starting ngrok and printer..."
	@trap 'kill 0' INT; \
	ngrok http --domain=ftdc-printer.ngrok.io --log=stdout 2>&1 8085 2>&1 & \
	go run cmd/printer/main.go 2>&1 & \
	wait

.PHONY: onemorething
onemorething:
	@echo "Starting ngrok and one more thing..."
	@trap 'kill 0' INT; \
	ngrok tcp --region=us --remote-addr=5.tcp.ngrok.io:27433 --log=stdout 2>&1 8588 2>&1 & \
	go run cmd/onemorething/main.go 2>&1 & \
	wait
