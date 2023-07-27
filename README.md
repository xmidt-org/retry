# retry

retry provides simple task retry logic.

[![Build Status](https://github.com/xmidt-org/retry/workflows/CI/badge.svg)](https://github.com/xmidt-org/retry/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/retry/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/retry?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/retry)](https://goreportcard.com/report/github.com/xmidt-org/retry)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/retry/blob/main/LICENSE)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_PROJECT&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_PROJECT)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/retry.svg)](CHANGELOG.md)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/xmidt-org/retry)](https://pkg.go.dev/github.com/xmidt-org/retry)


## Summary

retry provides simple ways of executing tasks with configurable retry semantics.  A focus is place on external configuration driving the retry behavior.  Tasks may be executed without retries, with a constant interval between retries, or using an exponential backoff.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Install](#install)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/docs/community/code_of_conduct/). 
By participating, you agree to this Code.

## Install

go get -u github.com/xmidt-org/retry

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
