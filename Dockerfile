FROM golang:1.24
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy


COPY * ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /share

EXPOSE 8080

CMD go run share web
# CMD ["/share"]
