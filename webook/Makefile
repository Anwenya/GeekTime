.PHONY: all
build:
	@rm webook || true
	@go mod tidy
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webook .
	@docker rmi -f wuliuqi/webook:v0.0.1
	@docker build -t wuliuqi/webook:v0.0.1 .
	@docker save -o webook-app.tar wuliuqi/webook:v0.0.1
	@crictl rmi wuliuqi/webook:v0.0.1
	@ctr -n k8s.io image import webook-app.tar
	@rm webook-app.tar && rm webook

k8s-apply-side:
	@kubectl apply -f webook-redis-deploy.yaml
	@kubectl apply -f webook-redis-service.yaml
	@kubectl apply -f webook-mysql-deploy.yaml
	@kubectl apply -f webook-mysql-service.yaml
	@kubectl apply -f webook-ingress.yaml

k8s-delete-side:
	@kubectl delete -f webook-redis-deploy.yaml
	@kubectl delete -f webook-redis-service.yaml
	@kubectl delete -f webook-mysql-deploy.yaml
	@kubectl delete -f webook-mysql-service.yaml
	@kubectl delete -f webook-ingress.yaml

k8s-apply-app:
	@kubectl apply -f webook-app-deploy.yaml
	@kubectl apply -f webook-app-service.yaml

k8s-delete-app:
	@kubectl delete -f webook-app-deploy.yaml
	@kubectl delete -f webook-app-service.yaml