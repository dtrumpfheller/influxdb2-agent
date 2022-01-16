#########################
# Build
#########################

FROM golang:1.17 as builder

WORKDIR /go/src/github.com/dtrumpfheller/influxdb2-agent

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go .
COPY helpers/*.go ./helpers/
COPY influxdb/*.go ./influxdb/

RUN CGO_ENABLED=0 go build -o /go/bin/app .


#########################
# Deploy
#########################

FROM gcr.io/distroless/static

COPY --from=builder /go/bin/app /

ENTRYPOINT ["/app"]