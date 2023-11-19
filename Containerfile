FROM docker.io/library/golang:1.21 as builder

WORKDIR /workdir
COPY go.mod go.sum main.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch

ENV USER_ID=1001
COPY --from=builder /workdir/promag /promag
EXPOSE 9000
WORKDIR /
USER ${USER_ID}
CMD ["/promag"]
