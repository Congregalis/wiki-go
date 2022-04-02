FROM golang:latest

WORKDIR /app/wiki-go
COPY . .

RUN go build wiki.go

EXPOSE 8080
ENTRYPOINT ["./wiki"]