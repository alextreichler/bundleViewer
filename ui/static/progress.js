function initProgress() {
    // Progress Bar Logic
    const progressBar = document.getElementById('progress-bar');
    const loadingOverlay = document.getElementById('loading-overlay');
    
    if (progressBar && loadingOverlay && !window.progressInitialized) {
        window.progressInitialized = true;
        const eventSource = new EventSource('/api/load-progress');

        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            const progress = data.progress;
            const status = data.status;
            const eta = data.eta;
            
            const loadingTitle = document.getElementById('loading-title');
            if (loadingTitle) {
                loadingTitle.textContent = `Analyzing Bundle... ${progress}%`;
            }

            progressBar.style.width = progress + '%';
            progressBar.textContent = progress + '%';
            
            const progressContainer = document.getElementById('progress-container');
            if (progressContainer) {
                progressContainer.style.display = 'block';
            }
            
            const statusEl = document.getElementById('progress-status');
            if (statusEl && status) {
                let displayStatus = status;
                if (eta > 0) {
                    const mins = Math.floor(eta / 60);
                    const secs = eta % 60;
                    const timeStr = mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
                    displayStatus += ` (ETA: ${timeStr})`;
                }
                statusEl.textContent = displayStatus;
            }

            if (progress >= 100) {
                eventSource.close();
                window.progressInitialized = false;
                setTimeout(() => {
                    loadingOverlay.style.display = 'none';
                }, 500);
            }
        };

        eventSource.onerror = function(err) {
            console.error("EventSource failed:", err);
            eventSource.close();
            window.progressInitialized = false;
            loadingOverlay.style.display = 'none';
        };
    }
}

if (document.readyState === 'loading') {
    document.addEventListener("DOMContentLoaded", function() {
        initProgress();
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initProgress();
        });
    });
} else {
    initProgress();
    if (!window.htmxProgressListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initProgress();
        });
        window.htmxProgressListenerAdded = true;
    }
}
