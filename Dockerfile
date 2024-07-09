FROM scratch

# Copy YugabyteDB Aeon CLI binary.
COPY ybm /ybm

# Set entry point.
ENTRYPOINT ["/ybm"]
