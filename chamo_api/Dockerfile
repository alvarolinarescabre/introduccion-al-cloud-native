FROM golang:1.24-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app


FROM golang:1.24-alpine
COPY --from=build /go/bin/app /
CMD ["/app"]
