# Kluster Controller - Self-Healing Deployment and Service

This project is a simple Kubernetes controller written in Go that watches a custom resource `Kluster` (defined under the API group `pulok.dev/v1alpha1`) and manages a Deployment and Service named `bookapi` in the `default` namespace accordingly.

It demonstrates:
- Watching custom resources with client-go informer
- Creating or recreating a Deployment and Service based on the custom resource spec
- Watching core Kubernetes Deployments and Services for deletion events and performing self-healing by recreating the deleted resources

---

## Features

- Watches for `Kluster` CR creation and automatically creates a Deployment and Service with the specified specs.
- Watches for deletion of the Deployment or Service labeled `app=bookapi` and automatically recreates them based on the existing `Kluster` resource.
- Default service port is 9090 if not specified.
- The Deployment runs a container with image and replicas configured via the `Kluster` custom resource.

---
