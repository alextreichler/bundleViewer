// static/theme.js
(function() {
    let currentTheme = localStorage.getItem('theme');
    // Handle legacy values
    if (currentTheme === 'dark-mode') currentTheme = 'dark';
    if (currentTheme === 'light-mode') currentTheme = 'light';

    const validThemes = ['light', 'dark', 'ultra-dark', 'compact', 'retro', 'powershell'];
    if (!validThemes.includes(currentTheme)) {
        currentTheme = 'light';
    }

    document.documentElement.setAttribute('data-theme', currentTheme);
})();

function initThemeToggle() {
    const displayBtn = document.querySelector('.theme-select-btn');
    if (!displayBtn) return;

    const themeOptions = document.querySelectorAll('.theme-select-option');

    function updateActiveTheme(theme) {
        // Update button text
        const themeName = theme.charAt(0).toUpperCase() + theme.slice(1).replace('-', ' ');
        displayBtn.textContent = `Themes: ${themeName}`;

        // Update active class on options
        themeOptions.forEach(opt => {
            if (opt.getAttribute('data-theme') === theme) {
                opt.classList.add('active-theme');
            } else {
                opt.classList.remove('active-theme');
            }
        });
    }

    // Initialize
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    updateActiveTheme(currentTheme);

    // Add listeners
    themeOptions.forEach(opt => {
        // Remove existing listener to prevent duplicates if called multiple times
        opt.onclick = null; 
        opt.addEventListener('click', function() {
            const selectedTheme = this.getAttribute('data-theme');
            document.documentElement.setAttribute('data-theme', selectedTheme);
            localStorage.setItem('theme', selectedTheme);
            updateActiveTheme(selectedTheme);
        });
    });
}

// Initial load trigger
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        initThemeToggle();
        // Add HTMX listener here to ensure body exists
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initThemeToggle();
        });
    });
} else {
    initThemeToggle();
    // Body definitely exists here
    if (!window.htmxThemeListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initThemeToggle();
        });
        window.htmxThemeListenerAdded = true;
    }
}
