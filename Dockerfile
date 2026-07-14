# Cross-compile from builder arch → TARGETOS/TARGETARCH (amd64 + arm64).
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/hookd ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/hookd /hookd
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/hookd"]
