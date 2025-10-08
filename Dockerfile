FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-metabase-v049"]
COPY baton-metabase-v049 /