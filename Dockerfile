FROM ubuntu:24.04

COPY /cmd/admission-webhook/webhook /app/webhook
WORKDIR /app
RUN chmod +x /app/webhook
ENTRYPOINT ["/app/webhook"]