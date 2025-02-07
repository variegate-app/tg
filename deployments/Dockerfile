## Builder
FROM golang:1.23-alpine3.19 AS dev-builder

ARG APPLICATION_NAME

# Create a workspace for the app
WORKDIR /app
ADD ../ ./

# Build
RUN go build -o ./application ./cmd/$APPLICATION_NAME

## Runner
FROM scratch AS dev-runner

WORKDIR /
COPY --from=dev-builder /app/application /application

ENTRYPOINT [ "/application" ]