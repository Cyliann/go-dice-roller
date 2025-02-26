FROM golang:1.23 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:latest
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
