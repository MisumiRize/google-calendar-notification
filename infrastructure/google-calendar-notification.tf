variable "aws_s3_bucket_name" {}
variable "google_client_id" {}
variable "google_client_secret" {}
variable "aws_region" {}
variable "twitter_consumer_key" {}
variable "twitter_consumer_secret" {}
variable "twitter_access_token" {}
variable "twitter_access_token_secret" {}

resource "aws_iam_user" "google_calendar_notification" {
  name = "google_calendar_notification"
}

resource "aws_iam_access_key" "google_calendar_notification" {
  user = "${aws_iam_user.google_calendar_notification.name}"
}

resource "aws_iam_user_policy" "google_calendar_notification_s3_readwrite" {
  name = "google_calendar_notification_s3_readwrite"
  user = "${aws_iam_user.google_calendar_notification.name}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1464923810000",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::${var.aws_s3_bucket_name}",
        "arn:aws:s3:::${var.aws_s3_bucket_name}/*"
      ]
    }
  ]
}
EOF
}

resource "heroku_app" "default" {
  name = "google-calendar-notification"
  region = "us"

  config_vars = {
    GOOGLE_CLIENT_ID = "${var.google_client_id}"
    GOOGLE_CLIENT_SECRET = "${var.google_client_secret}"
    AWS_ACCESS_KEY_ID = "${aws_iam_access_key.google_calendar_notification.id}"
    AWS_SECRET_ACCESS_KEY = "${aws_iam_access_key.google_calendar_notification.secret}"
    AWS_REGION = "${var.aws_region}"
    AWS_S3_BUCKET_NAME = "{$var.aws_s3_bucket_name}"
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
