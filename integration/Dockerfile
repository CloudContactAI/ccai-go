FROM golang:1.21-alpine AS build

WORKDIR /sdk
COPY go.mod go.sum ./
COPY src/ ./src/
COPY integration/ ./integration/

WORKDIR /sdk/integration
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /app/test .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /app/test /app/test
CMD ["/app/test"]
