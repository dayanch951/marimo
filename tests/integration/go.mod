module github.com/dayanch951/marimo/tests/integration

go 1.24.0

require (
	github.com/dayanch951/marimo/services/users v0.0.0
	github.com/dayanch951/marimo/shared v0.0.0
	github.com/gorilla/mux v1.8.1
)

replace (
	github.com/dayanch951/marimo/services/users => ../../services/users
	github.com/dayanch951/marimo/shared => ../../shared
)
