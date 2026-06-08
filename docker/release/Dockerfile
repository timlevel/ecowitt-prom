ARG VERSION=v1.0.0

FROM alpine:3.21
RUN apk --no-cache add ca-certificates curl
ARG VERSION
ARG TARGETARCH
RUN ARCH=${TARGETARCH:-amd64} && \
    curl -fsSL \
      "https://github.com/timlevel/ecowitt-prom/releases/download/${VERSION}/ecowitt-prom_linux_${ARCH}" \
      -o /usr/local/bin/ecowitt-prom && \
    chmod +x /usr/local/bin/ecowitt-prom
EXPOSE 8080
ENTRYPOINT ["ecowitt-prom"]
