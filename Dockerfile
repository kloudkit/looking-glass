# renovate: datasource=docker depName=golang
ARG go_tag=1.27.2

FROM golang:${go_tag} AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY internals ./internals
ARG version=dev
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X github.com/kloudkit/looking-glass/internals/cli.version=${version}" \
      -o /looking-glass .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /looking-glass /looking-glass
ENV PORT=8080
USER nonroot
ENTRYPOINT ["/looking-glass"]
