FROM gcr.io/distroless/base-debian12
COPY /changelog /
ENTRYPOINT ["/changelog"]
