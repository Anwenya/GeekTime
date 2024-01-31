.PHONY: build
docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -o webook .
	@docker rmi -f wuliuqi/webook:v0.0.1
	@docker build -t wuliuqi/webook:v0.0.1 .
	@docker save -o webook-app.tar wuliuqi/webook:v0.0.1
	@crictl rmi -f wuliuqi/webook:v0.0.1
	@ctr image import webook-app.tar
	@rm webook-app.tar & rm webook

