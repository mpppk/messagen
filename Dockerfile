FROM golang:1-buster AS builder
ENV GO111MODULE=on
RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 make build

FROM gcr.io/distroless/static-debian12
COPY --from=builder /src/messagen.bin /messagen
ENTRYPOINT ["/messagen"]
CMD ["run"]