FROM golang:1.12.4-stretch as base

COPY . /sbucket/
WORKDIR /sbucket

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o sbucket .

# This results in a single layer image
FROM scratch
COPY --from=base ./sbucket/sbucket ./sbucket

ENTRYPOINT ["/sbucket"]
