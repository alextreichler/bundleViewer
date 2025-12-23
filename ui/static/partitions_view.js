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

        let html = '<table><thead><tr><th>Partition ID</th><th>Leader</th><th>Replicas</th></tr></thead><tbody>';
        for (const partition of partitions) {
            html += `<tr><td>${partition.id}</td><td>${partition.leader}</td><td>${partition.replicas.join(', ')}</td></tr>`;
        }
        html += '</tbody></table>';

        detailRow.querySelector('td').innerHTML = html;
        detailRow.dataset.loaded = 'true';
    } catch (error) {
        console.error('Error loading partitions:', error);
        detailRow.querySelector('td').innerHTML = '<p>Error loading partitions. See console for details.</p>';
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

// Initialize after page load
document.addEventListener('DOMContentLoaded', () => {
    makePartitionTableSortable('#partitions-table');
    makePartitionTableSortable('#leaders-table');
    makePartitionTableSortable('#topics-table');
});
