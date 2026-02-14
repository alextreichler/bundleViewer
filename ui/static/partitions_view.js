async function toggleTopicDetails(topicName) {
    const detailRow = document.getElementById(topicName + '-details');
    
    if (detailRow.style.display === 'table-row') {
        detailRow.style.display = 'none';
        return;
    }

    // Check if already loaded
    if (detailRow.dataset.loaded === 'true') {
        detailRow.style.display = 'table-row';
        return;
    }

    // Show loading state
    detailRow.querySelector('td').innerHTML = '<p>Loading...</p>';
    detailRow.style.display = 'table-row';

    try {
        const response = await fetch(`/api/topic-details?topic=${encodeURIComponent(topicName)}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const partitions = await response.json();

        let html = '<table class="partitions-detail-table"><thead><tr><th>Partition ID</th><th>Status</th><th>Leader</th><th>Term</th><th>Replicas / Raft State</th><th>Logs</th><th>Actionable Insights</th></tr></thead><tbody>';
        for (const partition of partitions) {
            let statusClass = '';
            if (partition.status === 'under_replicated') statusClass = 'text-warning';
            if (partition.leader === -1) statusClass = 'insight-error';
            
            let raftStateHtml = partition.replicas.join(', ');
            if (partition.raft_details && partition.raft_details.length > 0) {
                raftStateHtml = '<ul style="margin:0; padding-left:15px; font-size:0.85em;">';
                for (const r of partition.raft_details) {
                    const leaderBadge = r.is_leader ? ' <span class="badge" style="background:#dcfce7; color:#166534;">Leader</span>' : '';
                    raftStateHtml += `<li>Node ${r.node_id}: Term ${r.term}${leaderBadge}</li>`;
                }
                raftStateHtml += '</ul>';
            }

            const ntp = `${partition.ns}/${partition.topic}/${partition.id}`;
            const logSearchUrl = `/logs?search="${encodeURIComponent(ntp)}"`;

            let insightHtml = '-';
            if (partition.insight) {
                const severityClass = partition.insight.severity === 'error' ? 'insight-error' : 'insight-warn';
                insightHtml = `
                    <div class="insight-box ${severityClass}">
                        <strong>${partition.insight.summary}</strong>
                        ${partition.insight.remediation ? `
                            <div class="remediation-cmd">
                                <code>${partition.insight.remediation}</code>
                                <button class="copy-btn" onclick="copyToClipboard('${partition.insight.remediation}')" title="Copy Command">📋</button>
                            </div>
                        ` : ''}
                    </div>
                `;
            }

            html += `<tr>
                <td>${partition.id}</td>
                <td><span class="${statusClass}">${partition.status || 'healthy'}</span></td>
                <td>${partition.leader === -1 ? '<span class="text-warning" style="font-weight:bold;">NONE</span>' : partition.leader}</td>
                <td>${partition.term}</td>
                <td>${raftStateHtml}</td>
                <td>
                    <div style="display:flex; gap:5px;">
                        <a href="${logSearchUrl}" class="button" style="padding: 2px 8px; font-size: 0.75rem;">🔍 Logs</a>
                        <button onclick="showRaftTimeline('${ntp}')" style="padding: 2px 8px; font-size: 0.75rem;">📜 Raft History</button>
                    </div>
                </td>
                <td>${insightHtml}</td>
            </tr>`;
        }
        html += '</tbody></table>';

        detailRow.querySelector('td').innerHTML = html;
        detailRow.dataset.loaded = 'true';
    } catch (error) {
        console.error('Error loading partitions:', error);
        detailRow.querySelector('td').innerHTML = '<p>Error loading partitions. See console for details.</p>';
    }
}

async function showRaftTimeline(ntp) {
    const modal = document.getElementById('log-context-modal');
    const body = document.getElementById('log-context-body');
    const header = modal.querySelector('h2');
    
    if (header) header.textContent = 'Raft Leadership History: ' + ntp;
    if (body) {
        body.innerHTML = '<p>Loading history from logs...</p>';
        modal.style.display = 'block';
    }

    try {
        const response = await fetch(`/api/raft/timeline?ntp=${encodeURIComponent(ntp)}`);
        const events = await response.json();

        if (events.length === 0) {
            body.innerHTML = '<p>No Raft-related events found in the available logs for this partition.</p>';
            return;
        }

        let html = '<div class="raft-timeline" style="padding: 10px;">';
        events.forEach(ev => {
            let color = 'var(--text-color)';
            let icon = 'ℹ️';
            let bg = 'rgba(0,0,0,0.02)';

            if (ev.type === 'Elected') {
                color = '#10b981'; // Green
                icon = '👑';
                bg = 'rgba(16, 185, 129, 0.05)';
            } else if (ev.type === 'Stepdown') {
                color = '#f59e0b'; // Orange
                icon = '📉';
                bg = 'rgba(245, 158, 11, 0.05)';
            } else if (ev.type === 'Timeout') {
                color = '#ef4444'; // Red
                icon = '⏰';
                bg = 'rgba(239, 68, 68, 0.05)';
            }

            html += `
                <div style="border-left: 3px solid ${color}; padding: 10px; margin-bottom: 10px; background: ${bg}; border-radius: 0 4px 4px 0;">
                    <div style="display: flex; justify-content: space-between; font-size: 0.8em; color: var(--text-color-muted); margin-bottom: 5px;">
                        <span>${new Date(ev.timestamp).toISOString().replace('T', ' ').substring(0, 23)}</span>
                        <span>Node: <strong>${ev.node}</strong></span>
                    </div>
                    <div style="font-weight: bold; color: ${color}; margin-bottom: 5px;">
                        ${icon} ${ev.type} ${ev.term > 0 ? `(Term ${ev.term})` : ''}
                        ${ev.reason ? `<span style="font-weight: normal; color: var(--text-color-muted); margin-left: 10px;">Reason: ${ev.reason}</span>` : ''}
                    </div>
                    <div style="font-family: var(--font-mono); font-size: 0.85em; word-break: break-all;">
                        ${ev.description}
                    </div>
                </div>
            `;
        });
        html += '</div>';
        body.innerHTML = html;
    } catch (error) {
        console.error('Error fetching raft timeline:', error);
        body.innerHTML = '<p>Error loading Raft history.</p>';
    }
}

function makePartitionTableSortable(tableId) {
    const table = document.querySelector(tableId);
    if (!table) return;

    const headers = table.querySelectorAll('th.sortable');
    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            sortPartitionTable(table, index, header);
        });
    });
}

function sortPartitionTable(table, column, header) {
    const tbody = table.querySelector('tbody');
    
    // Get all rows - we need to identify main vs detail rows more carefully
    const allRows = Array.from(tbody.children).filter(el => el.tagName === 'TR');
    
    // Build row pairs: main rows with their associated detail rows
    const rowPairs = [];
    for (let i = 0; i < allRows.length; i++) {
        const row = allRows[i];
        
        // Skip if this row is a detail row (has an id ending in -details)
        if (row.id && row.id.endsWith('-details')) {
            continue;
        }
        
        // This is a main row
        const pair = {
            mainRow: row,
            detailRow: null
        };
        
        // Check if next row is the detail row
        const nextRow = allRows[i + 1];
        if (nextRow && nextRow.id && nextRow.id.endsWith('-details')) {
            pair.detailRow = nextRow;
        }
        
        rowPairs.push(pair);
    }
    
    const isAsc = header.classList.contains('asc');

    // Remove sort classes from all headers in this table only
    table.querySelectorAll('th.sortable').forEach(h => h.classList.remove('asc', 'desc'));

    // Add appropriate class to clicked header
    header.classList.toggle('asc', !isAsc);
    header.classList.toggle('desc', isAsc);

    rowPairs.sort((a, b) => {
        // Get the cell content, excluding any nested links
        const getCellText = (row) => {
            const cell = row.cells[column];
            if (!cell) return '';
            // If cell contains a link, get the link text, otherwise get cell text
            const link = cell.querySelector('a');
            return link ? link.textContent.trim() : cell.textContent.trim();
        };

        const aCell = getCellText(a.mainRow);
        const bCell = getCellText(b.mainRow);

        // Try to parse as number (remove any non-numeric characters except . and -)
        const aNum = parseFloat(aCell.replace(/[^0-9.-]/g, ''));
        const bNum = parseFloat(bCell.replace(/[^0-9.-]/g, ''));

        // If both are valid numbers, sort numerically
        if (!isNaN(aNum) && !isNaN(bNum)) {
            return isAsc ? bNum - aNum : aNum - bNum;
        }

        // Otherwise sort alphabetically
        return isAsc ? bCell.localeCompare(aCell) : aCell.localeCompare(bCell);
    });

    // Clear tbody
    while (tbody.firstChild) {
        tbody.removeChild(tbody.firstChild);
    }

    // Re-append sorted rows with their detail rows
    rowPairs.forEach(pair => {
        tbody.appendChild(pair.mainRow);
        if (pair.detailRow) {
            tbody.appendChild(pair.detailRow);
        }
    });
}

function initPartitionsSortable() {
    makePartitionTableSortable('#partitions-table');
    makePartitionTableSortable('#leaders-table');
    makePartitionTableSortable('#topics-table');
}

// Initialize after page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        initPartitionsSortable();
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initPartitionsSortable();
        });
    });
} else {
    initPartitionsSortable();
    if (!window.htmxPartitionsListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initPartitionsSortable();
        });
        window.htmxPartitionsListenerAdded = true;
    }
}
