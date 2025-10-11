# PEP 0006: Add Parent Field for Monitors

Add a `parent` field to monitor resources in the Terraform provider for Uptime Kuma. This field will allow users to specify a parent monitor for hierarchical organization of monitors.
This has been missing in the previous implementation of the Go Uptime Kuma client library, but is now available.
