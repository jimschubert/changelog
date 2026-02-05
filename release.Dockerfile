FROM gcr.io/distroless/base-debian12:nonroot
COPY /changelog /
ENTRYPOINT ["/changelog"]
