FROM golang:1.12 AS build
WORKDIR /go/src/helloworld
COPY helloworld.go .
RUN go install .
RUN ls -al /go/bin/

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=build /go/bin/helloworld /app/helloworld
ENTRYPOINT ["/app/helloworld"]
