#include "metrics_client.h"

#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

static const int RECV_BUF_LEN = 1024;

static const int METRIC_JSON_MAX_LEN = 1024;

static char *const expected = "HTTP/1.1 200 OK";

int http_post(char *host, int port, char *method, char *path, char *postData, logger pImpl);

int make_socket(int timeout_seconds);

int connect_http(const char *host, int port, int socket_descriptor);

int post(int socket_descriptor, char *host, int port, char *method, char *path, char *postData);

int receive(int socket_descriptor);

int mk_metrics_json(char *dest, int max_len, char *service_name);

void send_otlp_metric(logger log, char *service_name) {
    char json[METRIC_JSON_MAX_LEN];
    int len = mk_metrics_json(json, METRIC_JSON_MAX_LEN, service_name);
    if (len == METRIC_JSON_MAX_LEN - 1) {
        log_debug(log, "otlp metric json too long, not sending");
        return;
    }
    char *host = "127.0.0.1";
    int port = 4318;
    char *method = "POST";
    char *path = "/v1/metrics";
    if (http_post(host, port, method, path, json, log)) {
        log_debug(log, "send otlp metric succeeded");
    } else {
        log_debug(log, "send otlp metric failed");
    }
}

int mk_metrics_json(char *dest, int max_len, char *service_name) {
    char *format = "{\"resourceMetrics\":[{\"resource\":{},\"scopeMetrics\":[{\"scope\":{},\"metrics\":"
                   "[{\"name\":\"splunk.linux-autoinstr.executions\",\"sum\":{\"dataPoints\":"
                   "[{\"attributes\":[{\"key\":\"service.name\",\"value\":{\"stringValue\":\"%s\"}}],\"asInt\":\"1\"}],"
                   "\"aggregationTemporality\":\"AGGREGATION_TEMPORALITY_DELTA\"}}]}]}]}";
    return snprintf(dest, max_len, format, service_name);
}

int http_post(char *host, int port, char *method, char *path, char *postData, logger log) {
    int socket_descriptor = make_socket(1);
    if (socket_descriptor == -1) {
        log_debug(log, "metrics client failed to open socket");
        return 0;
    }

    if (!connect_http(host, port, socket_descriptor)) {
        log_debug(log, "metrics client failed to connect");
        return 0;
    }

    if (!post(socket_descriptor, host, port, method, path, postData)) {
        log_debug(log, "metrics client failed to send");
        return 0;
    }

    if (!receive(socket_descriptor)) {
        log_debug(log, "metrics client failed to receive response");
        return 0;
    }

    return 1;
}

int make_socket(int timeout_seconds) {
    int socket_descriptor = socket(PF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (socket_descriptor == -1) {
        return -1;
    }

    struct timeval timeout;
    timeout.tv_sec = timeout_seconds;
    timeout.tv_usec = 0;

    if (setsockopt(socket_descriptor, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout)) < 0) {
        return -1;
    }

    if (setsockopt(socket_descriptor, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout)) < 0) {
        return -1;
    }

    return socket_descriptor;
}

int connect_http(const char *host, int port, int socket_descriptor) {
    struct sockaddr_in address;
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = inet_addr(host);
    address.sin_port = htons(port);
    int errno = connect(socket_descriptor, (struct sockaddr *) &address, sizeof(address));
    return errno == 0;
}

int post(int socket_descriptor, char *host, int port, char *method, char *path, char *postData) {
    char *req_pattern = "%s %s HTTP/1.1\n"
                        "Host: %s:%d\n"
                        "Content-Type: application/json\n"
                        "Content-Length: %d\n"
                        "User-Agent: splunk-zc/1.0\n"
                        "Accept: */*\n\n%s";
    char req_str[1024];
    sprintf(req_str, req_pattern, method, path, host, port, strlen(postData), postData);

    size_t req_len = strlen(req_str);
    ssize_t num_bytes_sent = send(socket_descriptor, req_str, req_len, 0);
    return num_bytes_sent == req_len;
}

int receive(int socket_descriptor) {
    char buf[RECV_BUF_LEN];
    ssize_t recv_size = recv(socket_descriptor, buf, RECV_BUF_LEN, 0);
    size_t expected_len = strlen(expected);
    if (recv_size < expected_len) {
        return 0;
    }

    return strncmp(expected, buf, expected_len) == 0;
}
