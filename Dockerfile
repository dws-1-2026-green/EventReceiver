FROM alpine:3.18
WORKDIR /app
COPY ./event-receiver /app/server
RUN chmod +x /app/server
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/server"]