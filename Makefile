APP_NAME := vboxes-clean-up

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
		-a -tags netgo \
		-o ${APP_NAME}-linux-amd64 \
