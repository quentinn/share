FROM golang:bookworm

WORKDIR /app
COPY go.mod *.go *.md ./
COPY templates/ ./templates/
COPY static/ ./static/


# - Download dependencies
# - Build
# - Force create the sqlite.db file to avoid app not start
RUN go mod tidy \
 && go build -o share \
 && ./share reset


EXPOSE 8080

CMD ["./share", "web"]