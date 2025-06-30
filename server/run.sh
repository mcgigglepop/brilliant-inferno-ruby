#!/bin/bash
go build -o application-build cmd/web/*.go && ./application-build  -cache=false -production=false -cognito-user-pool-id=1234 -cognito-client-id=1234