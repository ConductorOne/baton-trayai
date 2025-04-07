FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-trayai"]
COPY baton-trayai /