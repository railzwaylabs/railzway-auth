variable "version" {
  type        = string
  description = "Docker image tag for railzway-auth"
}

variable "github_token" {
  type        = string
  description = "GitHub PAT for pulling images from GHCR"
}

job "railzway-auth" {
  datacenters = ["dc1"]
  type        = "service"
  
  group "app" {
    count = 1

    network {
      port "http" {
        static = 8082
      }
    }

    service {
      name = "railzway-auth"
      port = "http"
      
      tags = [
        "traefik.enable=true",
        
        # HTTPS Router (main)
        "traefik.http.routers.railzway-auth.rule=Host(`auth.railzway.com`)",
        "traefik.http.routers.railzway-auth.entrypoints=websecure",
        "traefik.http.routers.railzway-auth.tls.certresolver=cloudflare",
        
        # HTTP Router (redirect to HTTPS)
        "traefik.http.routers.railzway-auth-http.rule=Host(`auth.railzway.com`)",
        "traefik.http.routers.railzway-auth-http.entrypoints=web",
        "traefik.http.routers.railzway-auth-http.middlewares=railzway-auth-redirect",
        
        # Redirect middleware
        "traefik.http.middlewares.railzway-auth-redirect.redirectscheme.scheme=https",
        "traefik.http.middlewares.railzway-auth-redirect.redirectscheme.permanent=true",
      ]

      check {
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "2s"
      }
    }

    # Application Server Task (handles migrations automatically on startup)
    task "server" {
      driver = "docker"

      config {
        image = "ghcr.io/railzwaylabs/railzway-auth:${var.version}"
        ports = ["http"]
        args  = ["serve"]
        
        # Docker registry authentication for GHCR
        auth {
          username = "railzwaylabs"
          password = var.github_token
        }
      }

      # Environment variables (non-sensitive)
      env {
        HTTP_PORT                 = "${NOMAD_PORT_http}"
        APP_ENV                   = "production"
        SERVICE_NAME              = "railzway-auth"
      }

      # Sensitive environment variables from Consul KV
      template {
        data = <<EOH
# Database
DATABASE_URL={{ key "railzway-auth/database_url" }}

# Redis
REDIS_ADDR={{ key "railzway-auth/redis_addr" }}
REDIS_PASSWORD={{ keyOrDefault "railzway-auth/redis_password" "" }}
REDIS_DB={{ keyOrDefault "railzway-auth/redis_db" "0" }}

# Bootstrap Admin
ADMIN_EMAIL={{ key "railzway-auth/admin_email" }}
ADMIN_PASSWORD={{ key "railzway-auth/admin_password" }}
DEFAULT_ORG={{ keyOrDefault "railzway-auth/default_org" "1000" }}

# Tokens
ACCESS_TOKEN_TTL={{ keyOrDefault "railzway-auth/access_token_ttl" "1h" }}
REFRESH_TOKEN_TTL={{ keyOrDefault "railzway-auth/refresh_token_ttl" "720h" }}
REFRESH_TOKEN_BYTES={{ keyOrDefault "railzway-auth/refresh_token_bytes" "32" }}

# Rate Limiting
RATE_LIMIT_RPM={{ keyOrDefault "railzway-auth/rate_limit_rpm" "600" }}

# CORS
CORS_ALLOWED_ORIGINS={{ keyOrDefault "railzway-auth/cors_allowed_origins" "*" }}
CORS_ALLOWED_METHODS={{ keyOrDefault "railzway-auth/cors_allowed_methods" "GET,POST,OPTIONS" }}
CORS_ALLOWED_HEADERS={{ keyOrDefault "railzway-auth/cors_allowed_headers" "Authorization,Content-Type" }}
CORS_ALLOW_CREDENTIALS={{ keyOrDefault "railzway-auth/cors_allow_credentials" "false" }}

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT={{ keyOrDefault "railzway-auth/otel_exporter_otlp_endpoint" "" }}
OTEL_EXPORTER_OTLP_INSECURE={{ keyOrDefault "railzway-auth/otel_exporter_otlp_insecure" "true" }}
EOH
        destination = "secrets/file.env"
        env         = true
      }

      resources {
        cpu    = 500  # MHz
        memory = 512  # MB
      }
    }

    # Rolling update strategy
    update {
      max_parallel     = 1
      min_healthy_time = "10s"
      healthy_deadline = "3m"
      auto_revert      = true
    }
  }
}
