FROM golang:bookworm

WORKDIR /share
COPY go.mod *.go *.md ./
COPY templates/ ./templates/
COPY static/ ./static/


# - Download dependencies
# - Build
# - Force create the sqlite.db file to avoid app not start
RUN go get -u \
 && go mod tidy \
 && go build -o share \
 && ./share reset


EXPOSE 8080

CMD ["./share", "web"]