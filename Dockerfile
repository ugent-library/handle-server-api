# build stage
FROM golang:1.19-alpine AS build
WORKDIR /build
COPY . .
RUN go get -d -v ./...
# avoid error "error obtaining VCS status: exec: "git": executable file not found in $PATH"
# by adding -buildvcs=false
RUN go build -buildvcs=false -o handle-server-api -v

# final stage
FROM alpine:latest

ARG SOURCE_BRANCH
ARG SOURCE_COMMIT
ARG IMAGE_NAME
ENV SOURCE_BRANCH $SOURCE_BRANCH
ENV SOURCE_COMMIT $SOURCE_COMMIT
ENV IMAGE_NAME $IMAGE_NAME

WORKDIR /dist

COPY --from=build /build/handle-server-api handle-server-api
EXPOSE 3000
CMD ["/dist/handle-server-api", "server"]
