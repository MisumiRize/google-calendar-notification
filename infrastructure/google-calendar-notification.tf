variable "twitter_consumer_key" {}
variable "twitter_consumer_secret" {}
variable "twitter_access_token" {}
variable "twitter_access_token_secret" {}

resource "heroku_app" "default" {
  name = "google-calendar-notification"
  region = "us"

  config_vars = {
    TWITTER_CONSUMER_KEY = "${var.twitter_consumer_key}"
    TWITTER_CONSUMER_SECRET = "${var.twitter_consumer_secret}"
    TWITTER_ACCESS_TOKEN = "${var.twitter_access_token}"
    TWITTER_ACCESS_TOKEN_SECRET = "${var.twitter_access_token_secret}"
  }
}

resource "heroku_addon" "scheduler" {
  app = "${heroku_app.default.name}"
  plan = "scheduler:standard"
}
