// static/tables.js

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    initTables();
});

// Initialize after HTMX content swap
document.addEventListener('htmx:afterSwap', function() {
    initTables();
});

function initTables() {
    const tables = document.querySelectorAll('table.sortable-table');
    tables.forEach(table => {
        // Prevent double initialization
        if (table.getAttribute('data-sort-initialized') === 'true') {
            return;
        }
        makeTableSortable(table);
        table.setAttribute('data-sort-initialized', 'true');
    });
}

function makeTableSortable(table) {
    const headers = table.querySelectorAll('th');
    headers.forEach((header, index) => {
        // Add sorting indicators css class if not present
        header.classList.add('sortable-header');
        header.style.cursor = 'pointer';
        header.style.userSelect = 'none';
        
        // Add a sort icon container if desired, or handle via CSS classes
        
        header.addEventListener('click', () => {
            const tbody = table.querySelector('tbody');
            // Select only main rows to start grouping
            const mainRows = Array.from(tbody.querySelectorAll('tr.main-row'));
            
            // Map main rows to their groups (main + details)
            const rowGroups = mainRows.map(mainRow => {
                const detailsId = mainRow.getAttribute('data-details-id');
                const detailsRow = detailsId ? document.getElementById(detailsId) : null;
                const details = [];
                if (detailsRow) {
                    details.push(detailsRow);
                }
                return { main: mainRow, details: details };
            });
            
            // Also need to handle rows that are neither main-row nor details-row (if any exist)
            // But for this specific table, we have strict classes. 
            // Ideally, we'd grab all non-details rows, but .main-row is safe for now given the template.

            // Determine current sort order
            const currentOrder = header.getAttribute('data-order');
            const isAscending = currentOrder === 'asc';
            const newOrder = isAscending ? 'desc' : 'asc';
            
            // Reset other headers
            headers.forEach(h => {
                h.removeAttribute('data-order');
                h.classList.remove('sorted-asc', 'sorted-desc');
            });

            // Set current header
            header.setAttribute('data-order', newOrder);
            header.classList.add(newOrder === 'asc' ? 'sorted-asc' : 'sorted-desc');

            // Sort groups based on the main row
            rowGroups.sort((groupA, groupB) => {
                const rowA = groupA.main;
                const rowB = groupB.main;

                const getCellValue = (row, idx) => {
                    const cell = row.cells[idx];
                    if (!cell) return "";
                    // Check data-value first, then innerText
                    if (cell.hasAttribute('data-value')) {
                        return cell.getAttribute('data-value');
                    }
                    return cell.innerText.trim();
                };

                const cellA = getCellValue(rowA, index);
                const cellB = getCellValue(rowB, index);
                
                // 1. Try numeric sort
                // We use a stricter check than parseFloat to avoid "123 abc" being parsed as 123
                const isNumA = !isNaN(cellA) && !isNaN(parseFloat(cellA));
                const isNumB = !isNaN(cellB) && !isNaN(parseFloat(cellB));

                if (isNumA && isNumB) {
                    const numA = parseFloat(cellA);
                    const numB = parseFloat(cellB);
                    return newOrder === 'asc' ? numA - numB : numB - numA;
                }
                
                // 2. String sort (case-insensitive)
                const strA = cellA.toString().toLowerCase();
                const strB = cellB.toString().toLowerCase();
                
                if (strA < strB) {
                    return newOrder === 'asc' ? -1 : 1;
                }
                if (strA > strB) {
                    return newOrder === 'asc' ? 1 : -1;
                }
                return 0;
            });

            // Re-append sorted rows and their details
            rowGroups.forEach(group => {
                tbody.appendChild(group.main);
                group.details.forEach(detailRow => tbody.appendChild(detailRow));
            });
        });
    });
}

