FROM golang:1.26-alpine AS build
WORKDIR /src
ARG VERSION=dev
ARG REVISION=unknown
ARG BUILD_DATE=unknown
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/vm75/virtual-art-gallery/internal/version.Version=${VERSION} -X github.com/vm75/virtual-art-gallery/internal/version.Revision=${REVISION} -X github.com/vm75/virtual-art-gallery/internal/version.BuildDate=${BUILD_DATE}" -o /out/gallery ./cmd/gallery

FROM alpine:3.22
ARG VERSION=dev
ARG REVISION=unknown
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.title="Virtual Art Gallery" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/vm75/virtual-art-gallery"
RUN addgroup -S gallery && adduser -S -G gallery -u 10001 gallery && mkdir /data && chown gallery:gallery /data
COPY --from=build /out/gallery /usr/local/bin/gallery
USER gallery
ENV GALLERY_LISTEN_ADDR=:8080 GALLERY_DATA_DIR=/data GALLERY_SECURE_COOKIES=false
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget -q -O - http://127.0.0.1:8080/healthz >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/gallery"]
