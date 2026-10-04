export GOTOOLCHAIN := local

.PHONY: build site check test serve match

## build: render the site, print both PDFs and the social images, run the ATS check
build:
	go run ./cmd/build

## site: render HTML only (fast, no Chrome); still checks the existing PDFs
site:
	go run ./cmd/build -pdf=false

## check: only run the ATS check on the committed PDFs
check:
	go run ./cmd/build -check

## test: data-framework tests (schema errors, visibility flags, dates)
test:
	go test ./...

## serve: preview on http://localhost:8080
serve:
	go run ./cmd/serve

## match: compare a job description with the resume, e.g. make match JD=job.txt IN=pt
match:
	go run ./cmd/build -match $(JD) $(if $(IN),-lang $(IN))
