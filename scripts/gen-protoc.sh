#!/usr/bin/env bash

protoc -I ./api/ -I ${GOPATH}/src \
    ./api/worker.v1.proto --go_out=plugins=grpc:pkg/api/v1/ --govalidators_out=pkg/api/v1/