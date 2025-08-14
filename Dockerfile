FROM docker.io/library/debian:stable AS builder

RUN apt-get update -y && \
    apt-get install -y git g++ swig python3-dev libpython3-dev pkg-config golang

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

ARG GIT_HASH
ARG EMBED_VERSION

COPY . .
RUN go mod tidy
RUN env CGO_CFLAGS='-O2' go build -v -x -ldflags="-s -w -X 'github.com/MythicAgents/forgescript/pkg/versioninfo.embedVersion=${EMBED_VERSION}' -X 'github.com/MythicAgents/forgescript/pkg/versioninfo.embedGitRevision=${GIT_HASH}'"

FROM docker.io/library/debian:stable AS runner

RUN apt-get update -y && \
    apt-get install -y python3-dev libpython3-dev

COPY --from=builder /usr/src/app/forgescript /usr/local/bin/forgescript

CMD ["/usr/local/bin/forgescript", "-runtime-dir", "/Mythic/runtime"]
