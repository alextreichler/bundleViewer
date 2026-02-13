// Common UI functions

// Toggle Visibility Function
function toggleVisibility(id, linkElement) {
    var el = document.getElementById(id);
    if (!el) return;

    var icon = null;
    if (linkElement) {
        icon = linkElement.querySelector('.toggle-icon');
    }

    // Determine appropriate display type
    var targetDisplay = 'block';
    if (el.tagName === 'TR') {
        targetDisplay = 'table-row';
    } else if (el.tagName === 'TD' || el.tagName === 'TH') {
        targetDisplay = 'table-cell';
    }

    if (el.style.display === 'none') {
        el.style.display = targetDisplay;
        if (icon) icon.classList.add('expanded');
        if (linkElement) linkElement.classList.add('expanded');
    } else {
        el.style.display = 'none';
        if (icon) icon.classList.remove('expanded');
        if (linkElement) linkElement.classList.remove('expanded');
    }
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        // Optional: show a brief "Copied!" message
        alert("Command copied to clipboard!");
    }).catch(err => {
        console.error('Failed to copy: ', err);
    });
}

// Deep Linking Logic
function initDeepLinking() {
    const urlParams = new URLSearchParams(window.location.search);
    const fileName = urlParams.get('file');
    const query = urlParams.get('q');
    
    if (fileName) {
        // Find the details element containing this file
        // We look for the summary span text
        const summaries = document.querySelectorAll('summary .file-name');
        for (let span of summaries) {
            if (span.textContent.includes(fileName)) {
                const details = span.closest('details');
                if (details) {
                    details.open = true;
                    
                    // Wait for expansion/rendering if needed
                    setTimeout(() => {
                        details.scrollIntoView({ behavior: 'smooth', block: 'start' });
                        
                        // If there is a query, try to find/highlight it
                        if (query) {
                           // Use window.find() as a simple native way to highlight/jump
                           // Note: window.find is non-standard but widely supported in desktop browsers
                           // We scope it to the details element if possible? No, window.find searches the whole page.
                           // But we just scrolled there.
                           
                           // Alternatively, we can use a primitive Highlighter on the content div
                           const contentDiv = details.querySelector('.file-list-content');
                           if (contentDiv) {
                               highlightAndScroll(contentDiv, query);
                           }
                        }
                    }, 100);
                }
                break;
            }
        }
    }
}

// Initial load trigger
if (document.readyState === 'loading') {
    document.addEventListener("DOMContentLoaded", function() {
        initDeepLinking();
        // Add HTMX listener
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initDeepLinking();
        });
    });
} else {
    initDeepLinking();
    if (!window.htmxCommonListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initDeepLinking();
        });
        window.htmxCommonListenerAdded = true;
    }
}

function highlightAndScroll(container, text) {
    if (!text) return;
    
    // Simple text node walker to find matches
    const treeWalker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, null, false);
    const nodeList = [];
    while(treeWalker.nextNode()) nodeList.push(treeWalker.currentNode);
    
    const lowerText = text.toLowerCase();
    
    for (let node of nodeList) {
        const val = node.nodeValue;
        const idx = val.toLowerCase().indexOf(lowerText);
        
        if (idx !== -1) {
            // Found a match!
            // Create a range to select/highlight
            // Since we can't easily "highlight" without modifying DOM and potentially breaking syntax highlighting or JSON structure,
            // we will just scroll this element into view and maybe flash it.
            
            const range = document.createRange();
            range.setStart(node, idx);
            range.setEnd(node, idx + text.length);
            
            const span = document.createElement('mark');
            span.style.backgroundColor = 'yellow';
            span.style.color = 'black';
            
            // Complex to wrap text node partially without breaking things, 
            // but for a "jump to" feature, just scrolling the parent element is often enough.
            const parent = node.parentElement;
            parent.scrollIntoView({ behavior: 'smooth', block: 'center' });
            parent.style.transition = 'background-color 0.5s';
            parent.style.backgroundColor = '#fffbeb'; // light yellow
            
            // Try to use Selection API to highlight for user
            const sel = window.getSelection();
            sel.removeAllRanges();
            sel.addRange(range);
            
            break; // Stop after first match
        }
    }
}


// Custom Multi-Select Dropdown Helper
class CustomMultiSelect {
    constructor(selectElement, placeholder = "Select options...") {
        this.selectElement = selectElement;
        this.placeholder = placeholder;
        this.selectedValues = new Set();
        
        // Initialize from existing select
        Array.from(this.selectElement.selectedOptions).forEach(opt => {
            this.selectedValues.add(opt.value);
        });

        // Hide original select
        this.selectElement.style.display = 'none';

        // Create UI elements
        this.container = document.createElement('div');
        this.container.className = 'custom-multiselect-container';
        
        this.button = document.createElement('div');
        this.button.className = 'custom-multiselect-button';
        this.updateButtonText();
        
        this.dropdown = document.createElement('div');
        this.dropdown.className = 'custom-multiselect-dropdown';
        
        // Build options
        const options = Array.from(this.selectElement.options);
        
        options.forEach(opt => {
            if (opt.value === "") return; // Skip placeholder/empty value options if any

            const optionDiv = document.createElement('div');
            optionDiv.className = 'custom-multiselect-option';
            
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = opt.value;
            checkbox.checked = this.selectedValues.has(opt.value);
            
            const label = document.createElement('span');
            label.textContent = opt.text;
            
            optionDiv.appendChild(checkbox);
            optionDiv.appendChild(label);
            
            // Toggle logic
            optionDiv.addEventListener('click', (e) => {
                // Prevent bubbling to button click if clicked directly
                if (e.target !== checkbox) {
                    checkbox.checked = !checkbox.checked;
                }
                
                if (checkbox.checked) {
                    this.selectedValues.add(opt.value);
                    opt.selected = true;
                } else {
                    this.selectedValues.delete(opt.value);
                    opt.selected = false;
                }
                this.updateButtonText();
                
                // Trigger change event on original select so other listeners know
                this.selectElement.dispatchEvent(new Event('change'));
            });

            this.dropdown.appendChild(optionDiv);
        });

        this.container.appendChild(this.button);
        this.container.appendChild(this.dropdown);
        
        // Insert after original select
        this.selectElement.parentNode.insertBefore(this.container, this.selectElement.nextSibling);

        // Event Listeners
        this.button.addEventListener('click', (e) => {
            // Only toggle if we didn't click a pill remove button
            if (!e.target.classList.contains('multiselect-pill-remove')) {
                e.stopPropagation();
                this.toggleDropdown();
            }
        });

        document.addEventListener('click', (e) => {
            if (!this.container.contains(e.target)) {
                this.closeDropdown();
            }
        });
    }

    toggleDropdown() {
        this.container.classList.toggle('open');
    }

    closeDropdown() {
        this.container.classList.remove('open');
    }

    deselect(value) {
        this.selectedValues.delete(value);
        
        // Update original select
        Array.from(this.selectElement.options).forEach(opt => {
             if (opt.value === value) opt.selected = false;
        });

        // Update Checkbox in dropdown
        const checkbox = this.dropdown.querySelector(`input[value="${CSS.escape(value)}"]`);
        if (checkbox) checkbox.checked = false;

        this.updateButtonText();
        this.selectElement.dispatchEvent(new Event('change'));
    }

    updateButtonText() {
        this.button.innerHTML = ''; // Clear content

        const selectedOptions = Array.from(this.selectElement.options)
            .filter(opt => this.selectedValues.has(opt.value));

        if (selectedOptions.length === 0) {
            this.button.textContent = this.placeholder;
            this.button.classList.add('placeholder');
            return;
        }

        this.button.classList.remove('placeholder');

        // Logic: Show up to 2 pills, then +N
        const maxDisplay = 2;
        const displayItems = selectedOptions.slice(0, maxDisplay);
        const remaining = selectedOptions.length - maxDisplay;

        displayItems.forEach(opt => {
            const pill = document.createElement('span');
            pill.className = 'multiselect-pill';
            
            const text = document.createElement('span');
            text.textContent = opt.text;
            text.style.maxWidth = '100px';
            text.style.overflow = 'hidden';
            text.style.textOverflow = 'ellipsis';
            
            const removeBtn = document.createElement('span');
            removeBtn.className = 'multiselect-pill-remove';
            removeBtn.innerHTML = '&times;';
            removeBtn.title = "Remove";
            removeBtn.onclick = (e) => {
                e.stopPropagation(); // Stop bubbling to button click
                this.deselect(opt.value);
            };

            pill.appendChild(text);
            pill.appendChild(removeBtn);
            this.button.appendChild(pill);
        });

        if (remaining > 0) {
            const morePill = document.createElement('span');
            morePill.className = 'multiselect-pill more-pill';
            morePill.textContent = `+${remaining}`;
            morePill.title = "Show all selected items";
            this.button.appendChild(morePill);
        }
    }
}
