resource "google_project_service" "secretmanager" {
  service = "secretmanager.googleapis.com"
  disable_dependent_services = false
}

resource "google_project_service" "run" {
  service = "run.googleapis.com"
  disable_dependent_services = false
}

resource "google_secret_manager_secret" "dogecho-slack" {
  secret_id = "dogecho-slack-verification"

  replication {
    automatic = true
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_service_account" "dogecho" {
  account_id   = "dogecho"
  display_name = "DogEcho Webhook"
}

resource "google_secret_manager_secret_iam_member" "my-app" {
  secret_id = google_secret_manager_secret.dogecho-slack.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.dogecho.email}"
}

resource "google_cloud_run_service" "dogecho" {
  name     = "dogecho-srv"
  location = "us-central1"

  template {
    spec {
      containers {
        image = "gcr.io/gcp-sm-demo-next/dogecho@sha256:879cb4f5687ef58540bcb1dde14b0e3bff6d87690aa5c9924d27fdc1a6ee231c"
        env {
          name  = "SECRET_RESOURCE_NAME"
          value = "${google_secret_manager_secret.dogecho-slack.id}/versions/latest"
        }
      }
      service_account_name = google_service_account.dogecho.email
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [google_project_service.run]
}

resource "google_cloud_run_service_iam_member" "allUsers" {
  service  = google_cloud_run_service.dogecho.name
  location = google_cloud_run_service.dogecho.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "webhook_url" {
  value = "${google_cloud_run_service.dogecho.status[0].url}"
}
