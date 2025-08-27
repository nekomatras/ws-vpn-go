# Buidl:  sudo docker build --network=host --rm -t myapp . && sudo docker image prune -f
# Export: sudo docker save -o myapp.tar myapp:latest
# Import: sudo docker load -i /tmp/myapp.tar
# Run:    sudo docker run --rm myapp:latest

FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o myapp .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/myapp /
CMD ["/myapp"]
