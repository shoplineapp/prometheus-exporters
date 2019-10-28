require 'sidekiq'
require 'sidekiq/prometheus/exporter'

Sidekiq.configure_client do |config|
  config.redis = { url: "redis://#{ENV['SIDEKIQ_REDIS_ENDPOINT']}/#{ENV['SIDEKIQ_REDIS_DATABASE']}" }
end

run Sidekiq::Prometheus::Exporter.to_app
