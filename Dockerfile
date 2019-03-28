FROM golang:1.12.1-stretch as base

COPY . /sbucket/src/
WORKDIR /sbucket/src

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o sbucket .

# This results in a single layer image
FROM scratch
COPY --from=base ./sbucket/src/sbucket ./sbucket

ENTRYPOINT ["/sbucket"]
