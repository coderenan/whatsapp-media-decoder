FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache ffmpeg

COPY . .

RUN go mod tidy

EXPOSE 8091

CMD ["go", "run", "./cmd/server"]