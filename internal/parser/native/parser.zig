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
