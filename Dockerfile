# syntax=docker/dockerfile:1

FROM golang:1.22.6
WORKDIR /app
COPY purple-check ./
COPY .env ./
EXPOSE 9990
CMD ["./purple-check"]
