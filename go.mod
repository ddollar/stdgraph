module github.com/ddollar/stdgraph

go 1.19

require (
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/graph-gophers/graphql-transport-ws v0.0.2
	github.com/pkg/errors v0.9.1
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
)

replace github.com/graph-gophers/graphql-go => github.com/ddollar/graphql-go v1.4.0-ddollar
