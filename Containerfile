FROM docker.io/library/golang:latest as build
WORKDIR /workdir
COPY go.mod go.sum main.go /workdir/

RUN go mod tidy
RUN go build -o promag .

FROM scratch

COPY --from=build --chmod=755 /workdir/promag /
EXPOSE 8080
CMD ["/promag"]
