#############################################
# Build
#############################################
FROM --platform=$BUILDPLATFORM golang:1.21-alpine as build

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git

WORKDIR /go/src/github.com/webdevops/azure-resourcemanager-exporter

# Dependencies
COPY go.mod go.sum .
RUN go mod download

# Compile
COPY . .
RUN make test
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

#############################################
# Test
#############################################
FROM gcr.io/distroless/static as test
USER 0:0
WORKDIR /app
COPY --from=build /go/src/github.com/webdevops/azure-resourcemanager-exporter/azure-resourcemanager-exporter .
RUN ["./azure-resourcemanager-exporter", "--help"]

#############################################
# final-static
#############################################
FROM gcr.io/distroless/static as final-static
ENV LOG_JSON=1
WORKDIR /
COPY --from=test /app .
USER 1000:1000
ENTRYPOINT ["/azure-resourcemanager-exporter"]
