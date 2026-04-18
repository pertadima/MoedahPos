#!/bin/bash
export PATH=$PATH:/usr/local/bin
/usr/local/bin/golangci-lint run ./... --timeout=20m > lint_issues.txt 2>&1
