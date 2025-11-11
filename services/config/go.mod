module github.com/dayanch951/marimo/services/config

go 1.21

replace github.com/dayanch951/marimo/shared => ../../shared

require (
	github.com/dayanch951/marimo/shared v0.0.0
	github.com/gorilla/mux v1.8.1
)

require github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
