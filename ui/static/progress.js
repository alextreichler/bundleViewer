document.addEventListener("DOMContentLoaded", function() {
    // Progress Bar Logic
    const progressBar = document.getElementById('progress-bar');
    const loadingOverlay = document.getElementById('loading-overlay');
    
    if (progressBar && loadingOverlay) {
        const eventSource = new EventSource('/api/load-progress');

        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            const progress = data.progress;
            const status = data.status;
            
            progressBar.style.width = progress + '%';
            progressBar.textContent = progress + '%';
            
            const statusEl = document.getElementById('loading-status');
            if (statusEl && status) {
                statusEl.textContent = status;
            }

            if (progress >= 100) {
                eventSource.close();
                setTimeout(() => {
                    loadingOverlay.style.display = 'none';
                }, 500);
            }
        };

        eventSource.onerror = function(err) {
            console.error("EventSource failed:", err);
            eventSource.close();
            loadingOverlay.style.display = 'none';
        };
    }
});