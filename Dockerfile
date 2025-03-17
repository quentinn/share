FROM golang:latest

WORKDIR /app
COPY *.go go.mod go.sum *.md ./
COPY templates/ ./templates/
COPY static/ ./static/

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /share

# Force create the sqlite.db file to avoid app not start
# This uses the pseudo CLI of Share
RUN go run share reset

EXPOSE 8080

CMD ["go", "run", "share", "web"]