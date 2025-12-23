// static/tables.js

document.addEventListener('DOMContentLoaded', function() {
    initTables();
});

function initTables() {
    const tables = document.querySelectorAll('table.sortable-table');
    tables.forEach(table => {
        makeTableSortable(table);
    });
}

function makeTableSortable(table) {
    const headers = table.querySelectorAll('th');
    headers.forEach((header, index) => {
        header.style.cursor = 'pointer';
        header.addEventListener('click', () => {
            const tbody = table.querySelector('tbody');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            const isAscending = header.getAttribute('data-order') === 'asc';
            
            // Reset other headers
            headers.forEach(h => {
                h.removeAttribute('data-order');
                h.classList.remove('sorted-asc', 'sorted-desc');
            });

            // Set current header
            header.setAttribute('data-order', isAscending ? 'desc' : 'asc');
            header.classList.add(isAscending ? 'sorted-desc' : 'sorted-asc');

            rows.sort((a, b) => {
                const cellA = a.cells[index].innerText.trim();
                const cellB = b.cells[index].innerText.trim();
                
                // Try numeric sort
                const numA = parseFloat(cellA);
                const numB = parseFloat(cellB);
                
                if (!isNaN(numA) && !isNaN(numB)) {
                    return isAscending ? numB - numA : numA - numB;
                }
                
                // Date sort (simple heuristic)
                const dateA = new Date(cellA);
                const dateB = new Date(cellB);
                if (!isNaN(dateA) && !isNaN(dateB)) {
                    return isAscending ? dateB - dateA : dateA - dateB;
                }

                // String sort
                return isAscending ? cellB.localeCompare(cellA) : cellA.localeCompare(cellB);
            });

            rows.forEach(row => tbody.appendChild(row));
        });
    });
}

