FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/siguo ./cmd/siguo

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/siguo /siguo
EXPOSE 8080
ENTRYPOINT ["/siguo"]
