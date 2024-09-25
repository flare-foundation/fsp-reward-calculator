# build executable
FROM golang:1.22 AS builder

WORKDIR /build

# Copy and download dependencies using go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . ./

# Build the applications
RUN go build -o /app/fsp-rewards-calculator fsp-rewards-calculator

FROM debian:latest AS execution

ARG deployment=flare
ARG type=voting

#RUN apt-get -y update && apt-get -y install curl

WORKDIR /app
COPY --from=builder /app/fsp-rewards-calculator .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["./fsp-rewards-calculator" ]
