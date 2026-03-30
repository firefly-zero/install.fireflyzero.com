FROM golang:1.26.1-alpine3.22 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o install-bin

FROM alpine:3.22 AS prod
COPY --from=build /app/install-bin /install-bin
CMD ["/install-bin"]
