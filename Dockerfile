# syntax=docker/dockerfile:1

FROM golang:1.27 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/taskflow ./cmd/taskflow && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/taskflow /out/migrate /app/
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/taskflow"]