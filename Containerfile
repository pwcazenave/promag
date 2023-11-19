FROM docker.io/library/golang:1.21 as builder
WORKDIR /workdir
COPY go.mod go.sum main.go /workdir/

RUN go mod tidy
RUN go build

FROM scratch

ENV USER_ID=1001
COPY --from=builder --chmod=0755 /workdir/promag /promag
EXPOSE 9000
WORKDIR /
USER ${USER_ID}
CMD ["/promag"]
