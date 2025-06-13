FROM golang:1.24.3-alpine3.20 AS builder

WORKDIR /src

COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# production env - 1 stage
FROM builder AS prod
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/runner ./...

# development env - 1 stage
FROM builder AS dev
RUN go install github.com/air-verse/air@latest
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b $(go env GOPATH)/bin v1.62.0


# production env - 2 stage
FROM scratch AS final-prod
COPY --from=prod /src/bin/runner /bin/runner
CMD ["/bin/runner"]

# development env - 2 stage
FROM golang:1.24.3-alpine3.20 AS final-dev
WORKDIR /app
COPY --from=dev /src /app
CMD ["air"]
