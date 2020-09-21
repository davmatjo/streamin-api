FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN go build -o streamin .

#############################

FROM alpine

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/streamin .

ENTRYPOINT ["/app/streamin"]