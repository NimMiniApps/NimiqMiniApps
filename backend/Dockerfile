FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget
COPY --from=build /bin/server /bin/server
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s \
  CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/bin/server"]
