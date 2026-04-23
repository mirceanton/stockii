FROM alpine:3.23

USER 1000:1000

WORKDIR /app
COPY stockii /app/
COPY templates/ /app/templates/
COPY static/ /app/static/

ENV STOCKII_CONFIG_PATH=/config/stockii.db
ENV STOCKII_DATA_PATH=/data
ENV STOCKII_TEMPLATES_PATH=/app/templates
ENV STOCKII_STATIC_PATH=/app/static
ENV STOCKII_PORT=8080

ENTRYPOINT ["/app/stockii"]
