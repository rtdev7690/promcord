FROM gcr.io/distroless/static
COPY promcord /
ENTRYPOINT ["/promcord"]
