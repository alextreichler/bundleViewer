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

// Back to top button logic
document.addEventListener("DOMContentLoaded", function() {
    var mybutton = document.getElementById("back-to-top");
    if (mybutton) {
        window.onscroll = function() {scrollFunction()};
        mybutton.addEventListener("click", function() {
            document.body.scrollTop = 0; // For Safari
            document.documentElement.scrollTop = 0; // For Chrome, Firefox, IE and Opera
        });
    }

    function scrollFunction() {
        if (!mybutton) return;
        if (document.body.scrollTop > 20 || document.documentElement.scrollTop > 20) {
            mybutton.style.display = "block";
        } else {
            mybutton.style.display = "none";
        }
    }
});

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
