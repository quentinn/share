FROM golang:1.24
WORKDIR /app

COPY *.go go.mod go.sum *.md ./
# COPY *.go ./
COPY templates/ ./templates/
COPY static/ ./static/
# COPY *.md ./
COPY sqlite.db ./

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /share

EXPOSE 8080

CMD ["/share"]
CMD ls -al
CMD go run share web
