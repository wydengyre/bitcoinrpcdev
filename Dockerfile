FROM golang:1.23.2

# This Dockerfile exists because MacOS doesn't let us run downloaded binaries.
# During development on Mac, this Dockerfile can be used to run createdb.

WORKDIR /app

COPY . .
RUN go build ./cmd/createdb
CMD ./createdb