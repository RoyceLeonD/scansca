FROM python:3.11-alpine

WORKDIR /app

# Install MkDocs and Material theme
RUN pip install mkdocs mkdocs-material

# Set content directory
VOLUME ["/content", "/config"]

# Expose port
EXPOSE 8000

# Start MkDocs
ENTRYPOINT ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000", "-f", "/config/mkdocs.yml"]