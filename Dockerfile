FROM golang:latest
WORKDIR /app

COPY *.go go.mod go.sum *.md ./
COPY templates/ ./templates/
COPY static/ ./static/

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /share

EXPOSE 8080

CMD ["/share"]
CMD go run share web
