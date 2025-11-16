FROM golang:alpine AS build

WORKDIR /opt
COPY go.mod go.sum ./
RUN go mod download

COPY internal ./internal
RUN mkdir ./dist \
	&& CGO_ENABLED=0 go build -o ./dist/server -a -v ./internal/server
RUN ls -la

FROM scratch
COPY --from=build /opt/dist/server /server
COPY LICENSE README.md /

ENTRYPOINT ["/server"]
