# Thanks to https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

ENV USER=gogeizhalsbot
ENV UID=10001

# Create user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Install git. Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/gogeizhalsbot/
COPY . .

# Fetch dependencies.
RUN go mod download

# Build the binary.
RUN go build -ldflags="-w -s" -o /go/bin/gogeizhalsbot $GOPATH/src/gogeizhalsbot/cmd
RUN chown -R "${USER}:${USER}" /go/bin/gogeizhalsbot

############################
# STEP 2 build a small image
############################
FROM alpine

WORKDIR /app

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable.
COPY --from=builder /go/bin/gogeizhalsbot /app/gogeizhalsbot
COPY ./config.sample.yml /app/config.yml

# Use an unprivileged user.
USER gogeizhalsbot:gogeizhalsbot

EXPOSE 8080

ENTRYPOINT ["/app/gogeizhalsbot"]
