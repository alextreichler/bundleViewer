const std = @import("std");
const mem = std.mem;

pub const LogLineInfo = extern struct {
    line_start: u32,
    line_len: u32,
    level_offset: u32,
    level_len: u32,
    ts_offset: u32,
    ts_len: u32,
};

export fn zig_parse_logs(data: [*]const u8, len: usize, count_out: *usize) ?[*]LogLineInfo {
    const allocator = std.heap.c_allocator;
    
    var lines = std.ArrayList(LogLineInfo){};
    errdefer lines.deinit(allocator);

    var offset: usize = 0;
    while (offset < len) {
        // Find end of line
        var line_end = offset;
        while (line_end < len and data[line_end] != 10) {
            line_end += 1;
        }

        const current_line_len = line_end - offset;
        if (current_line_len > 0) {
            var info = LogLineInfo{
                .line_start = @intCast(offset),
                .line_len = @intCast(current_line_len),
                .level_offset = 0,
                .level_len = 0,
                .ts_offset = 0,
                .ts_len = 0,
            };

            // Fast check for Redpanda LEVEL YYYY-MM-DD
            const line_data = data[offset..line_end];
            if (line_data.len > 25) {
                const space_idx = mem.indexOfScalar(u8, line_data, ' ');
                if (space_idx != null and space_idx.? <= 6) {
                    const level = line_data[0..space_idx.?];
                    if (is_valid_level(level)) {
                        info.level_offset = @intCast(offset);
                        info.level_len = @intCast(space_idx.?);
                        
                        var ts_start = space_idx.? + 1;
                        while (ts_start < line_data.len and line_data[ts_start] == ' ') {
                            ts_start += 1;
                        }
                        
                        if (ts_start + 10 < line_data.len) {
                            if (line_data[ts_start + 4] == '-' and line_data[ts_start + 7] == '-') {
                                info.ts_offset = @intCast(offset + ts_start);
                                var ts_end = ts_start;
                                while (ts_end < line_data.len and line_data[ts_end] != ' ' and line_data[ts_end] != '[') {
                                    ts_end += 1;
                                }
                                info.ts_len = @intCast(ts_end - ts_start);
                            }
                        }
                    }
                }
            }

            lines.append(allocator, info) catch return null;
        }

        offset = line_end + 1;
    }

    count_out.* = lines.items.len;
    const result = lines.toOwnedSlice(allocator) catch return null;
    return result.ptr;
}

fn is_valid_level(level: []const u8) bool {
    if (mem.eql(u8, level, "INFO")) return true;
    if (mem.eql(u8, level, "WARN")) return true;
    if (mem.eql(u8, level, "ERROR")) return true;
    if (mem.eql(u8, level, "DEBUG")) return true;
    if (mem.eql(u8, level, "TRACE")) return true;
    if (mem.eql(u8, level, "FATAL")) return true;
    return false;
}

export fn zig_free_result(lines: [*]LogLineInfo, count: usize) void {
    const allocator = std.heap.c_allocator;
    const slice = lines[0..count];
    allocator.free(slice);
}

const delimiters = " \t\n\r[]{}() ,:=;<>|\\/-#";

// Pre-computed bitmask for fast delimiter checks
const delim_map = blk: {
    var map = [_]bool{false} ** 256;
    for (delimiters) |c| {
        map[c] = true;
    }
    break :blk map;
};

inline fn is_delimiter(c: u8) bool {
    return delim_map[c];
}

// Fingerprinting
export fn zig_fingerprint(data: [*]const u8, len: usize, out_buf: [*]u8, out_len: usize) usize {
    var in_idx: usize = 0;
    var out_idx: usize = 0;

    const in_data = data[0..len];

    while (in_idx < len and out_idx < out_len) {
        const c = in_data[in_idx];

        // 1. Quoted Strings
        if (c == '"' or c == '\'') {
            const quote = c;
            const marker = "<STR>";
            if (out_idx + marker.len <= out_len) {
                mem.copyForwards(u8, out_buf[out_idx..out_len], marker);
                out_idx += marker.len;
            }
            in_idx += 1;
            // Use SIMD-accelerated indexOfScalar
            if (mem.indexOfScalar(u8, in_data[in_idx..], quote)) |quote_idx| {
                in_idx += quote_idx + 1;
            } else {
                in_idx = len;
            }
            continue;
        }

        // 2. Delimiters
        if (is_delimiter(c)) {
            out_buf[out_idx] = c;
            out_idx += 1;
            in_idx += 1;
            continue;
        }

        // 3. Token
        const start = in_idx;
        // Use SIMD-accelerated indexOfAny
        if (mem.indexOfAny(u8, in_data[in_idx..], delimiters)) |delim_idx| {
            in_idx += delim_idx;
        } else {
            in_idx = len;
        }
        const token = in_data[start..in_idx];

        if (get_variable_marker_simd(token)) |marker| {
            if (out_idx + marker.len <= out_len) {
                mem.copyForwards(u8, out_buf[out_idx..out_len], marker);
                out_idx += marker.len;
            }
        } else {
            const copy_len = @min(token.len, out_len - out_idx);
            mem.copyForwards(u8, out_buf[out_idx..out_len], token[0..copy_len]);
            out_idx += copy_len;
        }
    }

    return out_idx;
}

fn get_variable_marker_simd(s: []const u8) ?[]const u8 {
    if (s.len == 0) return null;

    // Fast Hex check
    if (s.len >= 2 and s[0] == '0' and (s[1] == 'x' or s[1] == 'X')) {
        return "<HEX>";
    }

    var has_digit = false;
    var is_numeric_or_dot = true;

    var i: usize = 0;
    // Process 16-byte chunks with SIMD
    while (i + 16 <= s.len) : (i += 16) {
        const chunk: @Vector(16, u8) = s[i..][0..16].*;
        const digit_mask = (chunk >= @as(@Vector(16, u8), @splat('0'))) & (chunk <= @as(@Vector(16, u8), @splat('9')));
        const dot_mask = chunk == @as(@Vector(16, u8), @splat('.'));
        
        if (@reduce(.Or, digit_mask)) has_digit = true;
        if (!@reduce(.And, digit_mask | dot_mask)) {
            is_numeric_or_dot = false;
            break;
        }
    }

    if (is_numeric_or_dot) {
        // Remainder
        while (i < s.len) : (i += 1) {
            const c = s[i];
            if (c >= '0' and c <= '9') {
                has_digit = true;
            } else if (c != '.') {
                is_numeric_or_dot = false;
                break;
            }
        }
    }

    if (!is_numeric_or_dot) {
        // We already know it's not purely numeric, but we still need to check for any digit
        // if we want to return <*> instead of nothing.
        if (!has_digit) {
            while (i < s.len) : (i += 1) {
                if (s[i] >= '0' and s[i] <= '9') {
                    has_digit = true;
                    break;
                }
            }
        }
    }

    if (is_numeric_or_dot and has_digit) return "<NUM>";
    if (has_digit) return "<*>";

    return null;
}
