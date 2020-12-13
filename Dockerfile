FROM golang:1.13-buster AS build
COPY . /src/github.com/arussellsaw/news
RUN cd /src/github.com/arussellsaw/news && CGO_ENABLED=0 go build -o news -mod=vendor

FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates nodejs nodejs-npm


COPY --from=build /src/github.com/arussellsaw/news/readability-server /app/readability-server

WORKDIR /app/readability-server

RUN npm install

WORKDIR /app

COPY --from=build /src/github.com/arussellsaw/news/news /app/
COPY --from=build /src/github.com/arussellsaw/news/static /app/static
COPY --from=build /src/github.com/arussellsaw/news/tmpl /app/tmpl
EXPOSE 8080
ENTRYPOINT /app/news