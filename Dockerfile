FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o myapp .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/myapp /
CMD ["/myapp", "-mode=server"]
