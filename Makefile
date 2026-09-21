include .env 

export PROJECT_ROOT=${shell pwd}

run:
	export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
		go run ${PROJECT_ROOT}/cmd/api/main.go

