# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY . .

RUN go mod download -x

ARG TARGETARCH

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /bin/server ./main.go

FROM alpine:latest AS final

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

ARG PORT
ARG VERSION
ENV VERSION=$VERSION

COPY --from=build /bin/server /bin/

EXPOSE ${PORT:-8000}

ENTRYPOINT [ "/bin/server", "serve" ]
