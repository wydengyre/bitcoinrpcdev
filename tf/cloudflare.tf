terraform {
  required_providers {
    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 3.0"
    }
  }
}

provider "cloudflare" {
  # token pulled from $CLOUDFLARE_API_TOKEN
}

variable "zone_id" {
  default = "c3ba69da7acc03a04a6d06da8d3aa23a"
}

variable "domain" {
  default = "bitcoinrpc.dev"
}

variable "account_id" {
  default = "5cf95ced729a19ed34a8e3010a66d80c"
}

variable "pages_project_name" {
  default = "bitcoinrpcdev-pages"
}

resource "cloudflare_zone_settings_override" "bitcoinrpcdev_zone_settings_override" {
  zone_id = var.zone_id
  settings {
    always_use_https          = "on"
    automatic_https_rewrites  = "on"
    http3                     = "on"
    # since our page is a static blob, this shouldn't be risky
    security_level              = "essentially_off"
    tls_1_3                     = "on"
  }
}

resource "cloudflare_zone_dnssec" "bitcoinrpcdev_zone_dnssec" {
  zone_id = var.zone_id
}

resource "cloudflare_record" "bitcoinrpcdev_record_apex" {
  zone_id = var.zone_id
  name = "bitcoinrpc.dev"
  value = "bitcoinrpcdev-pages.pages.dev"
  type = "CNAME"
}

resource "cloudflare_record" "bitcoinrpcdev_record_www_a" {
  zone_id = var.zone_id
  name = "www.bitcoinrpc.dev"
  # see https://developers.cloudflare.com/pages/how-to/www-redirect/
  value = "192.0.2.1"
  type = "A"

  proxied = true
}

resource "cloudflare_record" "bitcoinrpcdev_record_txt_google_site_verification" {
  zone_id = var.zone_id
  name = "bitcoinrpc.dev"
  # see https://developers.cloudflare.com/pages/how-to/www-redirect/
  value = "google-site-verification=57dcOfbqHHKTErswNiTlgFTw4dzaMDyk-jzNT3q0C-o"
  type = "TXT"
  ttl = 3600
}

resource "cloudflare_ruleset" "bitcoinrpcdev_ruleset_redirect_www_to_apex" {
  zone_id     = var.zone_id
  name        = "redirect-www-to-apex"
  description = "Redirect www to apex"
  kind        = "zone"
  phase       = "http_request_dynamic_redirect"

  rules {
    action = "redirect"
    action_parameters {
      from_value {
        status_code = 301
        target_url {
          value = "https://bitcoinrpc.dev"
        }
        preserve_query_string = false
      }
    }
    expression  = "(http.host eq \"www.bitcoinrpc.dev\")"
    description = "Redirect www to apex"
    enabled     = true
  }
}

resource "cloudflare_pages_project" "bitcoinrpcdev_pages_project" {
  account_id        = var.account_id
  name              = var.pages_project_name
  production_branch = "main"

  deployment_configs {
    preview {
      always_use_latest_compatibility_date = false
      compatibility_date                   = "2023-05-01"
      compatibility_flags                  = []
      d1_databases                         = {}
      durable_object_namespaces            = {}
      environment_variables                = {}
      fail_open                            = false
      kv_namespaces                        = {}
      r2_buckets                           = {}
      usage_model                          = "bundled"
    }
    production {
      always_use_latest_compatibility_date = false
      compatibility_date                   = "2023-05-01"
      compatibility_flags                  = []
      d1_databases                         = {}
      durable_object_namespaces            = {}
      environment_variables                = {}
      fail_open                            = false
      kv_namespaces                        = {}
      r2_buckets                           = {}
      usage_model                          = "bundled"
    }
  }
}

resource "cloudflare_pages_domain" "bitcoinrpcdev_pages_domain" {
  account_id = var.account_id
  project_name = var.pages_project_name
  domain = var.domain
}

resource "cloudflare_access_application" "bitcoinrpcdev_pages_subdomain" {
  account_id = var.account_id
  name = "bitcoinrpcdev pages subdomain"
  domain = "bitcoinrpcdev-pages.pages.dev"
  type = "self_hosted"
  session_duration = "24h"
}

resource "cloudflare_access_application" "bitcoinrpcdev_pages_subdomains" {
  account_id = var.account_id
  name = "bitcoinrpcdev pages subdomains"
  domain = "*.bitcoinrpcdev-pages.pages.dev"
  type = "self_hosted"
  session_duration = "24h"
}

resource "cloudflare_access_policy" "bitcoinrpcdev_pages_subdomain_access_policy" {
  application_id = cloudflare_access_application.bitcoinrpcdev_pages_subdomain.id
  account_id = var.account_id
  name = "bitcoinrpcdev pages subdomain access policy"
  precedence = 1
  decision = "allow"

  include {
    email = ["bitcoinrpc@wydengyre.com"]
  }
}

resource "cloudflare_access_policy" "bitcoinrpcdev_pages_subdomains_access_policy" {
  application_id = cloudflare_access_application.bitcoinrpcdev_pages_subdomains.id
  account_id = var.account_id
  name = "bitcoinrpcdev pages subdomain access policy"
  precedence = 1
  decision = "allow"

  include {
    email = ["bitcoinrpc@wydengyre.com"]
  }
}
