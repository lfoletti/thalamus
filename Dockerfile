# th-gateway, the Thalamus gateway. Build from the root of the repository:
#   docker build -t thalamus .
# Run with the rights file and the key files mounted in /etc/thalamus:
#   docker run -p 8787:8787 -v /srv/thalamus:/etc/thalamus thalamus
FROM golang:1.26 AS build
# The release passes its tag; a local build is a development build.
ARG VERSION=
WORKDIR /src/gateway
COPY gateway/go.mod gateway/go.sum ./
RUN go mod download
COPY gateway/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/lfoletti/thalamus/gateway/version.Tag=${VERSION}" -o /out/th-gateway ./cmd/th-gateway

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.source="https://github.com/lfoletti/thalamus" \
      org.opencontainers.image.description="th-gateway, the Thalamus gateway over CortexDB" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /out/th-gateway /usr/local/bin/th-gateway
WORKDIR /var/lib/thalamus
EXPOSE 8787
ENTRYPOINT ["/usr/local/bin/th-gateway"]
CMD ["serve", "/etc/thalamus/rights.yaml", "--listen", "0.0.0.0:8787", "--audit", "/var/lib/thalamus/audit.jsonl"]
