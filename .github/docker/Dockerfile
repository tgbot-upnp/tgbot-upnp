FROM --platform=$BUILDPLATFORM golang:alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o tgbot-upnp -trimpath -ldflags "-s -w" ./cmd/main.go

FROM --platform=$BUILDPLATFORM alpine:latest
WORKDIR /data
COPY --from=build --chmod=0755 /src/tgbot-upnp /usr/bin/tgbot-upnp
RUN set -ex \
	&& apk add --no-cache tzdata ca-certificates
ENV TELEGRAM_DATA_DIR=/data
ENTRYPOINT [ "/usr/bin/tgbot-upnp" ]
