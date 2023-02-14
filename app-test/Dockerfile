FROM golang:alpine3.16 AS build
LABEL stage=build
WORKDIR /app
COPY . ./
RUN apk add build-base
RUN go build main.go

FROM alpine:3.16 AS runner
WORKDIR /app
COPY --from=build /app/main /app/main
CMD ["/app/main"]