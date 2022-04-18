default: testacc

# Run acceptance tests
.PHONY: testacc
testacc-compose:
	docker-compose up -d
	WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 10m
	docker-compose down -v

testacc:
	minikube kubectl -- apply -f deployment.yaml
	minikube tunnel &> /dev/null & echo $$! > tunnel.PID
	timeout 30s bash -c 'until minikube kubectl -- get svc/wikijs --output=jsonpath='{.status.loadBalancer}' | grep "ingress"; do : ; done'
	timeout 30s bash -c 'until minikube service --url wikijs | grep "http"; do : ; done'
	WIKIJS_HOST=`minikube service --url wikijs` WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 10m
	minikube kubectl -- delete -f deployment.yaml
	if [ -a tunnel.PID ]; then \
		(kill $$(cat tunnel.PID) && rm tunnel.PID) || true; \
	fi;