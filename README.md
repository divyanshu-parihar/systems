# Systems

Systems programming experiments in Go, starting with a PostgreSQL request router called **Bouncer**.

## Bouncer

Bouncer implements a basic **PgBouncer-inspired strategy for routing SQL requests through a small pool of reusable PostgreSQL connections**. It is an early learning implementation that uses round-robin routing to distribute queries across the pool.

### How it works

1. Opens persistent PostgreSQL connections using `pgx` and stores them in a pool keyed by connection ID.
2. Starts a TCP server on port `8080`.
3. Handles each connected client in a separate goroutine.
4. Reads one SQL statement per newline from the client.
5. Selects a backend connection using `lastUsed % MAX_CONNECTIONS`, executes the statement, and advances the counter.
6. Returns `Success: <command tag>` on success, or the database error before closing the client connection.

The current startup configuration creates two backend connections with IDs `0` and `1`. Queries alternate between these connections, reusing them instead of opening a new database connection for each request.

### Run locally

You need Go compatible with the version declared in `bouncer/go.mod` (currently `1.26.3`), a reachable PostgreSQL database, and a TCP client such as `nc`.

Set the `dsn` constant in `bouncer/cmd/start.go` to your database connection string. Configuration currently lives in source; environment-variable configuration is not implemented. Avoid committing database credentials.

From the repository root, start the server:

```sh
make bouncer
```

Or run it directly:

```sh
go run ./bouncer
```

In another terminal, connect and send newline-terminated SQL:

```sh
nc localhost 8080
```

For example:

```sql
SELECT 1;
```

The response is a command tag, such as `Success: SELECT 1`, rather than the query's result rows.

### Current scope and limitations

- The client interface is a custom, newline-delimited TCP protocol. It does not implement the PostgreSQL wire protocol, so PostgreSQL clients such as `psql` cannot connect to it directly.
- Routing happens per SQL statement. There is no transaction or session affinity, so consecutive statements from one client can run on different backend connections.
- The pool is fixed at startup. Routing assumes all connection IDs from `0` through `MAX_CONNECTIONS - 1` are populated.
- Concurrent clients share the routing counter and backend connections without synchronization or exclusive checkout. Safe concurrent execution and connection scheduling are not implemented yet.
- `MAX_REQUESTS` is stored but not enforced. The current constructor call sets it to `10` and sets `MAX_CONNECTIONS` to `2`.
- Startup does not check errors returned when adding backend connections. Connection recovery, graceful shutdown, and backend connection cleanup are not implemented yet.

This demonstrates the core idea of reusing a limited set of database connections to route incoming work. Full PgBouncer compatibility and production pool management are outside the current implementation.

## Project layout

```text
bouncer/
  main.go           Application entry point
  cmd/start.go      Database configuration and server startup
  pg/bouncer.go     Pool, TCP server, and round-robin query routing
  pg/connection.go  PostgreSQL connection wrapper
Makefile            Command to run Bouncer
```
