FROM rg.fr-par.scw.cloud/vibioh/scratch

COPY cacert.pem /etc/ssl/cert.pem

ENTRYPOINT [ "/discord" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

COPY release/discord_${TARGETOS}_${TARGETARCH} /discord
COPY release/sweeper_${TARGETOS}_${TARGETARCH} /sweeper
