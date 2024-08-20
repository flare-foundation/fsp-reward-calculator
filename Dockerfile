# build executable
FROM golang:1.22 AS builder

WORKDIR /build

# Copy and download dependencies using go mod
COPY fdc-client/flare-common ./fdc-client/flare-common
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . ./

# Build the applications
RUN go build -o /app/fsc-rewards ftsov2-rewarding

FROM debian:latest AS execution

ARG deployment=flare
ARG type=voting

#RUN apt-get -y update && apt-get -y install curl

WORKDIR /app
COPY --from=builder /app/fsc-rewards .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["./fsc-rewards" ]
