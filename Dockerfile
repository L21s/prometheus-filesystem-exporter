FROM alpine:latest
EXPOSE 8080
ENTRYPOINT ["./prometheus-filesystem-exporter"]
ADD https://github.com/larscheid-schmitzhermes/prometheus-filesystem-exporter/releases/download/1.0.0/prometheus-filesystem-exporter prometheus-filesystem-exporter
RUN chmod +x prometheus-filesystem-exporter
