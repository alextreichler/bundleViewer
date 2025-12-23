// spinner.js
let spinnerCounter = 0; // To handle multiple potential spinners, though for this specific case it's one.

function showSpinner(element, spinnerClass = 'loading-spinner') {
    spinnerCounter++;
    if (!element) {
        console.warn("showSpinner called with null or undefined element.");
        return;
    }

    // Create spinner element
    const spinner = document.createElement('div');
    spinner.className = spinnerClass;
    spinner.id = 'active-spinner-' + spinnerCounter; // Unique ID

    // Basic spinner styling (can be enhanced with CSS)
    spinner.style.border = '4px solid #f3f3f3';
    spinner.style.borderTop = '4px solid #3498db';
    spinner.style.borderRadius = '50%';
    spinner.style.width = '12px';
    spinner.style.height = '12px';
    spinner.style.animation = 'spin 1s linear infinite';
    spinner.style.display = 'inline-block';
    spinner.style.marginLeft = '10px';
    spinner.style.verticalAlign = 'middle';

    // Add keyframes for spin animation if not already present
    if (!document.getElementById('spinner-keyframes')) {
        const styleSheet = document.createElement('style');
        styleSheet.id = 'spinner-keyframes';
        styleSheet.innerHTML = `
            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }
        `;
        document.head.appendChild(styleSheet);
    }
    
    // Hide the original content of the button/element and append the spinner
    element.originalContent = element.innerHTML; // Store original content
    element.innerHTML = '';
    element.appendChild(spinner);
    element.disabled = true; // Disable button while loading
}

function hideSpinner(element) {
    spinnerCounter--;
    if (!element) {
        console.warn("hideSpinner called with null or undefined element.");
        return;
    }

    const spinner = element.querySelector('[id^="active-spinner-"]');
    if (spinner) {
        spinner.remove();
        if (element.originalContent) {
            element.innerHTML = element.originalContent; // Restore original content
        }
    }
    element.disabled = false; // Re-enable button
}
