// ui/static/logs_view.js

if (typeof window.logsViewInitialized === 'undefined') {
    window.logsViewInitialized = false;
}

// Global State
window.allLogs = [];
window.isFetching = false;
window.hasMoreLogsServer = true;
window.serverOffset = 0;
window.serverLimit = 1000; // Fetch larger chunks for virtual scroll
window.rowHeight = 85; // Accommodate ~3 lines of wrapped text
window.lastScrollY = -1;
window.renderFrameRequestId = null;

// DOM Elements
window.logTableBody = null;
window.logTable = null;
window.loadingIndicator = null;
window.filterState = {};

if (!window.logsViewInitialized) {
    window.initLogsView = function() {
        window.logTableBody = document.querySelector('.log-table tbody');
        window.logTable = document.querySelector('.log-table');
        window.loadingIndicator = document.getElementById('loading-indicator');
        
        // Initial Fetch
        window.applyFilters(true);

        // Scroll Event Listener (Throttled via RequestAnimationFrame)
        window.addEventListener('scroll', () => {
            if (!window.renderFrameRequestId) {
                window.renderFrameRequestId = requestAnimationFrame(window.renderVirtualLogs);
            }
        });

        // Resize Listener
        window.addEventListener('resize', () => {
            window.rowHeight = window.estimateRowHeight();
            window.renderVirtualLogs();
        });

        // Infinite Scroll Trigger (Approaching bottom)
        const infiniteScrollObserver = new IntersectionObserver((entries) => {
            if (entries[0].isIntersecting && window.hasMoreLogsServer && !window.isFetching) {
                window.fetchLogs();
            }
        }, { rootMargin: '200px' });
        
        if (window.loadingIndicator) {
            infiniteScrollObserver.observe(window.loadingIndicator);
        }

        // Search Input
        const queryInput = document.getElementById('query-input');
        if (queryInput) {
            queryInput.addEventListener('keypress', function (e) {
                if (e.key === 'Enter') window.applyFilters();
            });
        }

        // Initialize Custom Multi-Selects
        const levelSelect = document.getElementById('level-select');
        if (levelSelect) {
            new CustomMultiSelect(levelSelect, "All Levels");
        }
        
        const componentSelect = document.getElementById('component-select');
        if (componentSelect) {
            new CustomMultiSelect(componentSelect, "All Components");
        }

        const nodeSelect = document.getElementById('node-select');
        if (nodeSelect) {
            new CustomMultiSelect(nodeSelect, "All Nodes");
        }

        // Analyze Logs Button
        const analyzeBtn = document.getElementById('analyze-logs-button');
        if (analyzeBtn) {
            analyzeBtn.addEventListener('click', function() {
                const indicator = document.getElementById('analyze-loading-indicator');
                if (indicator) indicator.style.display = 'inline';
                window.location.href = '/logs/analysis';
            });
        }
    };

    window.estimateRowHeight = function() {
        const firstRow = window.logTableBody ? window.logTableBody.querySelector('tr.log-row') : null;
        return firstRow ? firstRow.offsetHeight : (window.rowHeight || 45);
    };

    window.buildApiUrl = function(offset, limit) {
        const params = new URLSearchParams();
        for (const [key, value] of Object.entries(window.filterState)) {
            if (Array.isArray(value)) {
                value.forEach(v => params.append(key, v));
            } else if (value) {
                params.set(key, value);
            }
        }
        params.set('offset', offset);
        params.set('limit', limit);
        return '/api/logs?' + params.toString();
    };

    window.applyFilters = function(isInitial = false) {
        // Capture filter state
        const levelSelect = document.getElementById('level-select');
        const componentSelect = document.getElementById('component-select');
        const nodeSelect = document.getElementById('node-select');
        window.filterState = {
            search: document.getElementById('query-input')?.value || '',
            level: levelSelect ? Array.from(levelSelect.selectedOptions).map(o => o.value) : [],
            node: nodeSelect ? Array.from(nodeSelect.selectedOptions).map(o => o.value) : [],
            component: componentSelect ? Array.from(componentSelect.selectedOptions).map(o => o.value) : [],
            sort: document.getElementById('sort-select')?.value || 'desc',
            startTime: document.getElementById('start-time')?.value || '',
            endTime: document.getElementById('end-time')?.value || ''
        };

        // Reset State
        window.allLogs = [];
        window.serverOffset = 0;
        window.hasMoreLogsServer = true;
        window.lastScrollY = -1;
        
        if (window.logTableBody) window.logTableBody.innerHTML = '';
        window.scrollTo(0, 0);

        window.fetchLogs();
    };

    window.fetchLogs = async function() {
        if (window.isFetching || !window.hasMoreLogsServer) return;
        window.isFetching = true;
        
        // Show localized spinner in the loading indicator
        const indicator = window.loadingIndicator;
        if (indicator) {
            indicator.innerHTML = '<div id="local-spinner" class="htmx-indicator" style="display:inline-block; margin-right:10px;"></div> Loading more logs...';
            indicator.classList.add('htmx-request');
        }

        const url = window.buildApiUrl(window.serverOffset, window.serverLimit);

        try {
            const response = await fetch(url);
            const data = await response.json();

            if (data.logs && data.logs.length > 0) {
                // Pre-process logs (escape HTML, linkify) for faster rendering later
                const processed = data.logs.map(log => {
                    const escapedMsg = escapeHtml(log.message);
                    return {
                        ...log,
                        escapedMsg: escapedMsg,
                        linkedMsg: linkifyMessage(escapedMsg),
                        hasTrace: log.message.includes('\n'),
                        severityClass: log.level === 'ERROR' ? 'severity-error' : (log.level === 'WARN' ? 'severity-warn' : '')
                    };
                });

                window.allLogs.push(...processed);
                window.serverOffset += data.logs.length;
                window.hasMoreLogsServer = data.logs.length === window.serverLimit;
                
                // Update Total Count UI
                const displaySpan = document.getElementById('log-count-display');
                if (displaySpan) {
                    displaySpan.textContent = `Loaded ${window.allLogs.length} of ${data.total} matching logs`;
                }

                // If first load, measure row height after render
                const isFirstLoad = window.allLogs.length === data.logs.length;
                window.renderVirtualLogs(true);
                
                if (isFirstLoad) {
                    setTimeout(() => {
                        window.rowHeight = window.estimateRowHeight();
                        window.renderVirtualLogs(true); // Re-render with correct height
                    }, 50);
                }

            } else {
                window.hasMoreLogsServer = false;
                if (window.allLogs.length === 0 && window.logTableBody) {
                    window.logTableBody.innerHTML = '<tr><td colspan="7"><div class="empty-state"><i>🔍</i>No logs found matching your criteria. Try broadening your filters.</div></td></tr>';
                }
            }
        } catch (error) {
            console.error("Fetch error:", error);
            if (indicator) indicator.textContent = 'Error loading logs.';
        } finally {
            window.isFetching = false;
            window.renderFrameRequestId = null;
            if (indicator) {
                 indicator.classList.remove('htmx-request');
                 indicator.textContent = window.hasMoreLogsServer ? 'Scroll for more...' : 'End of logs.';
            }
        }
    };

    window.renderVirtualLogs = function(force = false) {
        window.renderFrameRequestId = null;

        if (!window.logTable || !window.logTableBody || window.allLogs.length === 0) return;

        const scrollY = window.scrollY;
        
        // Optimization: Don't re-render if scroll hasn't changed enough to shift rows
        // unless forced (e.g. new data arrived)
        if (!force && Math.abs(scrollY - window.lastScrollY) < (window.rowHeight / 2)) {
           return;
        }
        window.lastScrollY = scrollY;

        // Calculate Visible Window
        // We assume the table starts after the header. 
        // The header is sticky, but the *document* scroll position determines where we are in the "list".
        // The table's top position relative to the document:
        const tableTop = window.logTable.offsetTop + window.logTable.tHead.offsetHeight;
        const viewportHeight = window.innerHeight;
        
        // Relative scroll position within the data list
        const relativeScroll = Math.max(0, scrollY - tableTop);
        
        const buffer = 10; // Extra rows top/bottom
        const startIndex = Math.max(0, Math.floor(relativeScroll / window.rowHeight) - buffer);
        const endIndex = Math.min(window.allLogs.length, Math.ceil((relativeScroll + viewportHeight) / window.rowHeight) + buffer);

        // Calculate Padding
        const paddingTop = startIndex * window.rowHeight;
        const paddingBottom = (window.allLogs.length - endIndex) * window.rowHeight;

        // Render
        const fragment = document.createDocumentFragment();

        // Top Spacer
        if (paddingTop > 0) {
            const tr = document.createElement('tr');
            tr.style.height = `${paddingTop}px`;
            tr.style.border = 'none'; // Prevent borders on spacers
            // Needs a cell to be valid HTML, but we can hide it or make it empty
            const td = document.createElement('td');
            td.colSpan = 7;
            td.style.padding = 0;
            td.style.border = 'none';
            tr.appendChild(td);
            fragment.appendChild(tr);
        }

        // Data Rows
        const searchInput = document.getElementById('query-input');
        const searchTerms = getSearchTerms(searchInput ? searchInput.value : '');

        for (let i = startIndex; i < endIndex; i++) {
            const log = window.allLogs[i];
            const tr = document.createElement('tr');
            tr.className = 'log-row ' + (log.severityClass || '');
            tr.style.height = `${window.rowHeight}px`; // Enforce height
            
            // View Trace Button
            let actionBtn = '';
            if (log.hasTrace) {
                actionBtn = `<button class="trace-btn" onclick="openTraceModal(${i})" style="font-size: 0.8em; padding: 2px 6px; cursor: pointer;">View Trace</button>`;
            }

            // Insight Badge
            let insightHtml = '';
            if (log.insight) {
                insightHtml = `<div class="insight-badge severity-${log.insight.severity}" title="${log.insight.action}">
                    ${log.insight.description}
                </div>`;
            }

            // Message Display (Truncated if too long for one line, but we rely on CSS mostly)
            // We use a div inside the cell to enforce max-height/ellipsis if needed
            // Updated to allow 3 lines of wrapping by default, expandable on click
            const messageHtml = `
                <div class="log-message-container" onclick="this.classList.toggle('expanded')">
                    ${insightHtml}
                    ${log.linkedMsg}
                </div>`;

            const pinBtnClass = log.isPinned ? 'pin-btn pinned' : 'pin-btn';
            const pinBtnText = log.isPinned ? '★' : '☆';

            tr.innerHTML = `
                <td class="col-timestamp">${new Date(log.timestamp).toISOString().replace('T', ' ').substring(0, 23)}</td>
                <td class="log-level-${log.level} col-level">${log.level}</td>
                <td class="col-node">${log.node}</td>
                <td class="col-shard">${log.shard}</td>
                <td>${log.component}</td>
                <td class="log-message-cell">${messageHtml}</td>
                <td style="width: 120px;">
                    <div style="display: flex; gap: 5px; align-items: center;">
                        <button class="${pinBtnClass}" onclick="togglePin(${i})" title="${log.isPinned ? 'Unpin' : 'Pin to Notebook'}">${pinBtnText}</button>
                        <button onclick="showLogContext('${log.filePath}', ${log.lineNumber})" style="font-size: 0.8em; padding: 2px 6px;">Context</button>
                        <a href="/metrics?t=${log.timestamp}" class="button" style="font-size: 0.8em; padding: 2px 6px; background-color: var(--primary-color);">Metrics</a>
                        ${actionBtn}
                    </div>
                </td>
            `;

            // Highlighting (Performance intensive, strictly limit to visible rows)
            if (searchTerms.length > 0) {
                const msgDiv = tr.cells[5].firstElementChild;
                if(msgDiv) highlightElement(msgDiv, searchTerms);
            }

            fragment.appendChild(tr);
        }

        // Bottom Spacer
        if (paddingBottom > 0) {
            const tr = document.createElement('tr');
            tr.style.height = `${paddingBottom}px`;
            tr.style.border = 'none';
            const td = document.createElement('td');
            td.colSpan = 7;
            td.style.padding = 0;
            td.style.border = 'none';
            tr.appendChild(td);
            fragment.appendChild(tr);
        }

        window.logTableBody.innerHTML = '';
        window.logTableBody.appendChild(fragment);
    };

    // --- Helper Functions ---
    
    window.openTraceModal = function(index) {
        const log = window.allLogs[index];
        if (!log) return;
        
        const modal = document.getElementById('log-context-modal'); // Reuse context modal
        const body = document.getElementById('log-context-body');
        const header = modal.querySelector('h2');
        
        if (header) header.textContent = 'Log Trace';
        if (body) {
            body.innerHTML = `
                <div style="margin-bottom: 10px; font-weight: bold;">${log.level} @ ${log.timestamp}</div>
                <div style="background: var(--table-alt-row-bg); padding: 10px; border-radius: 4px; overflow-x: auto;">
                    ${log.escapedMsg} 
                </div>
            `;
        }
        modal.style.display = 'block';
    };

    window.logsViewInitialized = true;
}

// Helpers (Keep existing ones mostly)
function escapeHtml(text) {
    if (!text) return '';
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}

function linkifyMessage(text) {
    // Basic regex for K8s resources - same as before
    const regex = /(pod|node|service|pvc|deployment|statefulset)[s]?([\/: ]\s*)([a-z0-9](?:[-a-z0-9]*[a-z0-9])?)/gi;
    return text.replace(regex, (match, type, sep, name) => {
        let idPrefix = type.toLowerCase();
        if (idPrefix.endsWith('s')) idPrefix = idPrefix.slice(0, -1);
        const map = { 'pod': 'pod-', 'node': 'node-', 'service': 'service-', 'pvc': 'pvc-', 'deployment': 'deploy-', 'statefulset': 'sts-' };
        const prefix = map[idPrefix] || (idPrefix + '-');
        return `${type}${sep}<a href="/k8s#${prefix}${name}" target="_blank" style="text-decoration: underline; color: var(--link-color);">${name}</a>`;
    });
}

function getSearchTerms(query) {
    if (!query) return [];
    const tokens = query.match(/"[^ vital ]+"|\S+/g) || [];
    const terms = [];
    const operators = new Set(['AND', 'OR', 'NOT', '(', ')']);
    tokens.forEach(token => {
        let term = token;
        if (term.startsWith('"') && term.endsWith('"') && term.length >= 2) term = term.substring(1, term.length - 1);
        if (operators.has(term.toUpperCase())) return;
        const colonIdx = term.indexOf(':');
        if (colonIdx > 0) {
            // Very basic field support for highlighting only
            const value = term.substring(colonIdx + 1);
            terms.push(value);
        } else {
            terms.push(term);
        }
    });
    return terms.filter(t => t.length > 0);
}

function highlightElement(element, terms) {
    if (!terms || terms.length === 0) return;
    try {
        const distinctTerms = [...new Set(terms)].sort((a, b) => b.length - a.length);
        if (distinctTerms.length === 0) return;
        
        // Simple regex escape
        const pattern = new RegExp(`(${distinctTerms.map(t => t.replace(/[.*+?^${}()|[\\]/g, '\\$&')).join('|')})`, 'gi');
        
        // We only highlight text nodes to preserve HTML (links)
        const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT, null, false);
        const nodes = [];
        while(walker.nextNode()) nodes.push(walker.currentNode);
        
        nodes.forEach(node => {
             if (node.parentNode.tagName === 'MARK' || node.parentNode.tagName === 'A') return; // Skip already highlighted or links
             const text = node.nodeValue;
             if (pattern.test(text)) {
                 const span = document.createElement('span');
                 span.innerHTML = escapeHtml(text).replace(pattern, '<mark>$1</mark>');
                 node.parentNode.replaceChild(span, node);
             }
        });
    } catch(e) { console.error("Highlight error", e); }
}

// Re-expose legacy functions for buttons
function clearFilters() {
    window.location.reload();
}

async function showLogContext(filePath, lineNumber) {
    const modal = document.getElementById('log-context-modal');
    const contextBody = document.getElementById('log-context-body');
    const header = modal.querySelector('h2');
    if(header) header.textContent = 'Log Context';
    
    if (!contextBody) return;
    contextBody.textContent = 'Loading...';
    if (modal) modal.style.display = 'block';
    
    try {
        const safeFilePath = encodeURIComponent(filePath);
        const url = `/api/logs/context?filePath=${safeFilePath}&lineNumber=${lineNumber}&contextSize=10`;
        const response = await fetch(url);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const contextLines = await response.json();
        contextBody.innerHTML = '';
        contextLines.forEach((line, index) => {
            const lineElem = document.createElement('div');
            lineElem.textContent = line;
            lineElem.className = 'log-context-line';
            if (index % 2 === 0) lineElem.classList.add('even-line'); else lineElem.classList.add('odd-line');
            if (index === 10) { lineElem.style.backgroundColor = 'var(--highlight-bg)'; lineElem.style.fontWeight = 'bold'; }
            contextBody.appendChild(lineElem);
        });
    } catch (error) {
        contextBody.textContent = `Error loading log context: ${error.message}`;
    }
}

async function togglePin(index) {
    const log = window.allLogs[index];
    if (!log) return;

    const method = log.isPinned ? 'unpin' : 'pin';
    const url = `/api/logs/${method}?id=${log.id}`;

    try {
        const response = await fetch(url, { method: 'POST' });
        if (response.ok) {
            log.isPinned = !log.isPinned;
            window.renderVirtualLogs(true);
        } else {
            console.error(`Failed to ${method} log`);
        }
    } catch (error) {
        console.error(`Error during ${method}:`, error);
    }
}

// Init
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        if (window.initLogsView) window.initLogsView();
        // Add HTMX listener
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            if (document.querySelector('.log-table')) {
                if (window.initLogsView) window.initLogsView();
            }
        });
    });
} else if (typeof window.initLogsView === 'function') {
    window.initLogsView();
    if (!window.htmxLogsListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            if (document.querySelector('.log-table')) {
                if (window.initLogsView) window.initLogsView();
            }
        });
        window.htmxLogsListenerAdded = true;
    }
}