# API Server Documentation

## Getting Started

This guide provides instructions on how to run and test the API server. The server is built in Go and utilizes standard libraries to handle RESTful operations.

### Running the Server

To start the server, execute the following command in your terminal:

```bash
go run main.go
```


### to check if the port is being used, use the command
```bash
lsof -i :8080
```
to kill the server use command
```bash
kill -9 33364
```

example commands to test the server:

retrieve all books
```bash
curl -X GET http://localhost:8080/books \
    -H "X-API-Key: secret-key"
```

retrieve a single book
```bash
curl -X GET http://localhost:8080/book/1 \
    -H "X-API-Key: secret-key"
```

update an existing book
```bash
curl -X PUT http://localhost:8080/book/1 \
    -H "Content-Type: application/json" \
    -H "X-API-Key: secret-key" \
    -d '{"id": "1", "title": "The Great Gatsby Revised"}'
```

delete a book
```bash
curl -X DELETE http://localhost:8080/book/1 \
    -H "X-API-Key: secret-key"
```