[splunk-connect-for-otlp://default]
interval = 0
grpc_port = <4317>
http_port = <4318>
listen_address = <0.0.0.0>
realm = <us0>
access_token = <access token>
enableSSL = 1
serverCert = <location of server certificate file>
serverKey = <location of server key file>

# HEC token refresh settings. All values must be positive durations with units,
# such as 30s, 3m, or 1h. Omitted settings use the defaults below.
refresh_interval = <duration>
# How often to fetch a complete HEC token snapshot. Default: 3m.
refresh_timeout = <duration>
# Deadline for fetching all token pages. Default: 30s.
max_staleness = <duration>
# Maximum accepted token snapshot age. Default: 10m.
# Must be greater than refresh_interval + refresh_timeout.
