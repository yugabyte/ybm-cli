FROM scratch

# Copy YBM CLI binary.
COPY ybm /ybm

# Set entry point.
ENTRYPOINT ["/ybm"]
