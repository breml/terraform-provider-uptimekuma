# PEP 0004: Better Default Values for Accepted Status Codes

Empty list for `accepted_status_codes` in HTTP monitor is almost never meaningful, since it will match no responses and therefore always trigger a failure.
This field should be optional and default to `["200-299"]`.
