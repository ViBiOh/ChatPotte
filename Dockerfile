FROM rg.fr-par.scw.cloud/vibioh/scratch

COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT [ "/discord" ]

ARG VERSION
ENV VERSION ${VERSION}

ARG GIT_SHA
ENV GIT_SHA ${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

COPY release/discord_${TARGETOS}_${TARGETARCH} /discord
COPY release/sweeper_${TARGETOS}_${TARGETARCH} /sweeper
