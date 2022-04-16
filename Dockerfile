FROM vibioh/scratch

COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT [ "/discord_configure" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY release/discord_configure_${TARGETOS}_${TARGETARCH} /discord_configure
