//go:generate protoc -I . --go_out=. --go-grpc_out=. --grpc-gateway_out=. sports/sports.proto

package proto