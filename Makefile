# Run all tests.
# set COVERAGE_DIR If not set
COVERAGE_DIR ?= .coverage
.PHONY: test
test:
	@echo "[go test] running unit tests and collecting coverage metrics"
	@-rm -r $(COVERAGE_DIR)
	@mkdir $(COVERAGE_DIR)
	@go test -v -race -covermode atomic -coverprofile $(COVERAGE_DIR)/combined.txt ./...

.PHONY: test_update
test_update:
	@echo "[go test] running tests with updating golden file and collecting coverage metrics"
	@-rm -r $(COVERAGE_DIR)
	@mkdir $(COVERAGE_DIR)
	@go test -v -update -race -covermode atomic -coverprofile $(COVERAGE_DIR)/combined.txt ./...

# get the html coverage
html-coverage:
	@go tool cover -html=$(COVERAGE_DIR)/combined.txt
