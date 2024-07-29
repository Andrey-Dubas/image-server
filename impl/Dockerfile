FROM golang:1.22 as builder

WORKDIR /app
COPY * ./
RUN sleep 1 && echo ls
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./...

FROM alpine
WORKDIR /app
COPY --from=builder /app/app ./app
EXPOSE 8080
VOLUME /app_images C:\\app_images
CMD [ "/app/app" ]