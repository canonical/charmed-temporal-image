temporal-server:
	@printf $(COLOR) "Build temporal-server with CGO_ENABLED=$(CGO_ENABLED) for $(GOOS)/$(GOARCH)..."
	go build -o temporal-server .

temporal-cassandra-tool:
	@printf $(COLOR) "Build temporal-cassandra-tool with CGO_ENABLED=$(CGO_ENABLED) for $(GOOS)/$(GOARCH)..."
	go build -o temporal-cassandra-tool ./cmd/tools/cassandra

temporal-sql-tool:
	@printf $(COLOR) "Build temporal-sql-tool with CGO_ENABLED=$(CGO_ENABLED) for $(GOOS)/$(GOARCH)..."
	go build -o temporal-sql-tool ./cmd/tools/sql
