document.body.addEventListener('htmx:afterSettle', function(evt) {
    if (document.getElementById('timelineChart')) {
        loadTimelineChart();
    }
});

async function loadTimelineChart() {
    const canvas = document.getElementById('timelineChart');
    if (!canvas) return;

    try {
        const response = await fetch('/api/timeline/aggregated?bucket=2m');
        if (!response.ok) throw new Error('Failed to fetch aggregated data');
        
        const data = await response.json();
        renderTimelineChart(canvas, data);
    } catch (e) {
        console.error("Chart load error:", e);
        canvas.parentNode.innerHTML = '<div style="text-align:center; padding: 20px; color: var(--text-color-muted);">Failed to load timeline chart.</div>';
    }
}

function renderTimelineChart(canvas, data) {
    const ctx = canvas.getContext('2d');
    
    // Destroy existing chart if present to prevent "Canvas is already in use" error
    const existingChart = Chart.getChart(canvas);
    if (existingChart) {
        existingChart.destroy();
    }
    
    // Theme colors
    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || 
                       document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';
    const barColor = isDarkMode ? 'rgba(255, 99, 132, 0.9)' : 'rgba(255, 99, 132, 0.8)';

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.map(d => new Date(d.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})),
            datasets: [{
                label: 'Errors & Warnings',
                data: data.map(d => d.count),
                backgroundColor: barColor,
                borderColor: isDarkMode ? 'rgba(255, 99, 132, 1)' : 'rgba(255, 99, 132, 0.8)',
                borderWidth: 1,
                barPercentage: 0.9,      // Restore small gap between bars
                categoryPercentage: 0.9, // Restore small gap between categories
                minBarLength: 8,         // Keep 1-event visibility
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                title: { display: true, text: 'Event Frequency (per 2 minutes)', color: textColor }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: { color: gridColor },
                    ticks: { color: textColor, precision: 0 }
                },
                x: {
                    grid: { display: false },
                    ticks: { 
                        color: textColor
                    }
                }
            },
            onClick: (e, elements) => {
                if (elements.length > 0) {
                    const idx = elements[0].index;
                    const point = data[idx];
                    // On click, navigate to logs view with this time range
                    const start = new Date(point.timestamp);
                    const end = new Date(start.getTime() + 120000); // +2 mins
                    
                    // Format for input type="datetime-local" (YYYY-MM-DDTHH:MM)
                    const fmt = d => d.toISOString().substring(0, 16);
                    window.location.href = `/logs?level=ERROR&level=WARN&startTime=${fmt(start)}&endTime=${fmt(end)}`;
                }
            }
        }
    });
}