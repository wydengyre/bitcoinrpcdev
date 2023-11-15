FROM golang:1.21.4

WORKDIR /app

COPY . .
# TODO: maybe this can just be cross-compiled on Mac?
RUN go build ./cmd/createdb
CMD ./createdb