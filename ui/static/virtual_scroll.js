// ui/static/virtual_scroll.js
async function initVirtualScroll() {
    const tableContainer = document.getElementById('partition-leaders-container');
    const tbody = document.getElementById('partition-leaders-tbody');

    if (!tableContainer || !tbody) {
        return;
    }

    if (tableContainer.dataset.vscrollInitialized === 'true') {
        return;
    }
    tableContainer.dataset.vscrollInitialized = 'true';

    // Helper to handle 0 values correctly (so they don't turn into dashes)
    const safeGet = (val, fallback = '-') => (val !== undefined && val !== null) ? val : fallback;

    // 1. Fetch Data
    let data;
    try {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center;">Loading partition leaders...</td></tr>';
        
        const response = await fetch('/api/partition-leaders');
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        data = await response.json();

    } catch (error) {
        console.error("Failed to load partition leaders:", error);
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: red;">Failed to load partition leaders.</td></tr>';
        tableContainer.dataset.vscrollInitialized = 'false';
        return;
    }

    if (!data || data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center;">No partition leaders found.</td></tr>';
        return;
    }

    // 2. Setup Virtual Scroll Constants
    const rowHeight = 35; // Set to match your CSS row height
    let visibleRows = Math.ceil(tableContainer.clientHeight / rowHeight);
    const bufferRows = 5;

    // Reset styles that might conflict
    tbody.style.display = '';
    tbody.style.position = '';
    
    let lastStartIndex = -1;

    const renderRows = () => {
        const scrollTop = tableContainer.scrollTop;
        
        // Calculate the window of data to show
        const startIndex = Math.max(0, Math.floor(scrollTop / rowHeight) - bufferRows);
        const endIndex = Math.min(data.length, startIndex + visibleRows + (2 * bufferRows));

        if (startIndex === lastStartIndex) return;
        lastStartIndex = startIndex;

        // Calculate size of invisible top and bottom areas
        const paddingTop = startIndex * rowHeight;
        const paddingBottom = (data.length - endIndex) * rowHeight;

        tbody.innerHTML = '';
        const fragment = document.createDocumentFragment();

        // A. Top Spacer Row (Pushes content down)
        if (paddingTop > 0) {
            const topRow = document.createElement('tr');
            topRow.style.height = `${paddingTop}px`;
            // Empty cell is required to maintain table structure
            const cell = document.createElement('td');
            cell.colSpan = 7; 
            cell.style.padding = 0;
            cell.style.border = 'none';
            topRow.appendChild(cell);
            fragment.appendChild(topRow);
        }

        // B. Render Visible Data Rows
        for (let i = startIndex; i < endIndex; i++) {
            const leader = data[i];
            const row = document.createElement('tr');
            row.style.height = `${rowHeight}px`;

            // Data Mapping with Fallbacks
            const cells = [
                safeGet(leader.ns || leader.namespace),
                safeGet(leader.topic),
                safeGet(leader.partition_id || leader.partition),
                safeGet(leader.leader || leader.leader_id),
                safeGet(leader.term ?? leader.last_stable_leader_term),
                safeGet(leader.revision ?? leader.rev ?? leader.partition_revision),
                safeGet(leader.previous_leader ?? leader.previous_leader_id)
            ];

            cells.forEach(content => {
                const cell = document.createElement('td');
                cell.textContent = content;
                cell.style.whiteSpace = 'nowrap';
                cell.style.overflow = 'hidden';
                cell.style.textOverflow = 'ellipsis';
                cell.style.padding = '8px';
                row.appendChild(cell);
            });

            fragment.appendChild(row);
        }

        // C. Bottom Spacer Row (Maintains scrollbar size)
        if (paddingBottom > 0) {
            const bottomRow = document.createElement('tr');
            bottomRow.style.height = `${paddingBottom}px`;
            const cell = document.createElement('td');
            cell.colSpan = 7;
            cell.style.padding = 0;
            cell.style.border = 'none';
            bottomRow.appendChild(cell);
            fragment.appendChild(bottomRow);
        }

        tbody.appendChild(fragment);
    };

    // Initial render
    renderRows();

    // Event Listeners
    tableContainer.addEventListener('scroll', () => {
        window.requestAnimationFrame(renderRows);
    });

    const resizeObserver = new ResizeObserver(() => {
        visibleRows = Math.ceil(tableContainer.clientHeight / rowHeight);
        lastStartIndex = -1; // Force re-render
        renderRows();
    });
    resizeObserver.observe(tableContainer);

    // Disable sorting headers (since virtual scroll breaks standard sorting)
    const table = tbody.closest('table');
    if (table) {
        table.classList.remove('sortable-table');
        const headers = table.querySelectorAll('th');
        headers.forEach(th => {
            th.removeAttribute('onclick');
            th.style.cursor = 'default';
            th.classList.remove('sorted-asc', 'sorted-desc');
        });
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        initVirtualScroll();
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initVirtualScroll();
        });
    });
} else {
    initVirtualScroll();
    if (!window.htmxVScrollListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initVirtualScroll();
        });
        window.htmxVScrollListenerAdded = true;
    }
}
