# build stage
FROM golang:alpine AS build
WORKDIR /build
COPY . .
RUN go get -d -v ./...
# avoid error "error obtaining VCS status: exec: "git": executable file not found in $PATH"
# by adding -buildvcs=false
RUN go build -buildvcs=false -o handle-server-api -v

# final stage
FROM alpine:latest
WORKDIR /dist
COPY --from=build /build .
EXPOSE 3000
CMD "/dist/handle-server-api"
