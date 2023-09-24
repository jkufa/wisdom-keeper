# Use the official golang base image
FROM golang:latest

WORKDIR /usr/bin

COPY . .

RUN go build -o wisdom-keeper

# Set the entry point
CMD ["./wisdom-keeper"]
