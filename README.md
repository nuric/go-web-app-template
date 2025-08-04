# go-api-template

This is a template for creating RESTful APIs in Go using the standard library as the main framework. It is designed to be simple and minimal so you can easily extend it to your needs. It includes:

- Standard library HTTP server with routing
- Middleware using HTTP handlers including recovery and logging
- Standard library testing for handlers
- JSON encoding and decoding
- For convenience, logging and environment variable based configuration
- Sample http requests for testing in the [samples.http](samples.http) file
- 2-stage Dockerfile for easy deployment with slim builds

**Why?** When I start projects, I often have to scaffold a lot of boilerplate code. People argue that's what frameworks are for, but often I need something that's customised down the line. The goal of this template is to provide that initial start with minimal framework overhead.

## Getting Started

You can use this template to create a new Go project. Select use this template in Github to get started. To run the server:

```bash
go run ./
```

or using Docker:

```bash
docker build -t go-api-template .
docker run -p 8080:8080 go-api-template
```

Then use the [samples.http](samples.http) file to test the API which works with [Rest Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension. You can of course use any other HTTP client like Postman or curl.

## TODO

- Add password reset functionality
- Implement user profile management
- Rate limit

## Built with

- [env](https://github.com/caarlos0/env) for environment variable based configuration
- [zerolog](https://github.com/rs/zerolog) for logging
- [testify](https://github.com/stretchr/testify) for testing
