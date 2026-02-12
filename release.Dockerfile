FROM gcr.io/distroless/base-debian12:nonroot
ARG APP_NAME
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/${APP_NAME} /changelog
ENTRYPOINT ["/changelog"]
