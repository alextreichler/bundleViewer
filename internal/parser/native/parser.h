#ifndef LOG_PARSER_H
#define LOG_PARSER_H

#include <stddef.h>
#include <stdint.h>

typedef struct {
    uint32_t line_start;
    uint32_t line_len;
    uint32_t level_offset;
    uint32_t level_len;
    uint32_t ts_offset;
    uint32_t ts_len;
} LogLineInfo;

typedef struct {
    LogLineInfo* lines;
    size_t count;
    size_t capacity;
} ParseResult;

LogLineInfo* zig_parse_logs(const char* data, size_t len, size_t* count_out);
void zig_free_result(LogLineInfo* lines, size_t count);

// Fingerprinting
size_t zig_fingerprint(const char* data, size_t len, char* out_buf, size_t out_len);

// Metrics Parsing
typedef struct {
    uint32_t name_offset;
    uint32_t name_len;
    uint32_t labels_offset;
    uint32_t labels_len;
    double value;
} MetricInfo;

MetricInfo* zig_parse_metrics(const char* data, size_t len, size_t* count_out);
void zig_free_metrics(MetricInfo* metrics, size_t count);

#endif
