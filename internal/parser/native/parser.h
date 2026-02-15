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

#endif
