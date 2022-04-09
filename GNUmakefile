default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	docker-compose up -d
	#WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks go test ./wikijs/... -v $(TESTARGS) -timeout 120m
	#TF_LOG_SDK_HELPER_RESOURCE=debug TF_LOG_SDK_HELPER_SCHEMA=debug TF_LOG=debug WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
	TF_LOG_SDK_HELPER_RESOURCE=debug TF_LOG_SDK_HELPER_SCHEMA=debug TF_ACC_STATE_LINEAGE=1 TF_LOG=debug WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./internal/provider/... -v $(TESTARGS) -timeout 120m
	#WIKIJS_HOST=http://localhost:8080 WIKIJS_USERNAME=admin@wiki.example.local WIKIJS_PASSWORD=wikijsrocks TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
	docker-compose down -v
