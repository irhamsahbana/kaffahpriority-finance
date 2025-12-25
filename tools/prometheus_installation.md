docker rm -f prometheus-kaffah || true
docker run -d \
  --name prometheus-kaffah \
  -p 9191:9191 \
  -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --web.listen-address=:9191