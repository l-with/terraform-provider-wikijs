default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	docker-compose up -d
	WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 10m
	docker-compose down -v
