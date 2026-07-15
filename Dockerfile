FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/hookd /hookd
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/hookd"]
