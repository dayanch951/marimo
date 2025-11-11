module github.com/dayanch951/marimo/services/users

go 1.21

replace github.com/dayanch951/marimo/shared => ../../shared

require (
	github.com/dayanch951/marimo/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	google.golang.org/grpc v1.60.1
)
