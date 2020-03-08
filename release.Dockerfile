FROM gcr.io/distroless/base-debian10
COPY /changelog /
ENTRYPOINT ["/changelog"]
