
FROM golang:1.20.2-alpine3.17
ARG GOPRIVATE="github.com/okpalaChidiebere/*"
ENV GOPRIVATE $GOPRIVATE
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN --mount=type=secret,id=gitcredentials,target=/root/.netrc \
  CGO_ENABLED=0 go build -a -installsuffix cgo -o main main.go

FROM alpine:latest 
LABEL maintainer="Chidiebere Okpala <okpalacollins4@gmail.com>"
WORKDIR /app
COPY --from=0 /app/main .
EXPOSE 9000

# Run stage
CMD [ "/app/main" ]