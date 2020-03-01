FROM golang:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM gcr.io/distroless/static
ENV HTTP_ADDRESS ":8080"
LABEL maintainer="Reiner Gerecke <me@reinergerecke.de>"
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["/main"] 