// metrics_view.js - Global chart initialization functions

window.explorerChart = null;

// Helper to get timestamp from URL
function getTargetTimestamp() {
    const params = new URLSearchParams(window.location.search);
    const t = params.get('t');
    if (!t) return null;
    return new Date(t).getTime();
}

// Vertical line plugin for Chart.js
const verticalLinePlugin = {
    id: 'verticalLine',
    afterDraw: (chart, args, opts) => {
        if (!opts.timestamp) return;
        
        const timestamp = opts.timestamp;
        const xAxis = chart.scales.x;
        const yAxis = chart.scales.y;
        
        // Find x position. Works for both linear (numeric ms) and category (string) scales
        let xPos;
        if (xAxis.type === 'linear') {
            if (timestamp < xAxis.min || timestamp > xAxis.max) return;
            xPos = xAxis.getPixelForValue(timestamp);
        } else {
            // For category scales (SAR charts), we need to find the closest index
            // SAR labels are HH:mm:ss
            const targetStr = new Date(timestamp).toISOString().substring(11, 19);
            if (!chart.data.labels) return;
            const index = chart.data.labels.indexOf(targetStr);
            if (index === -1) {
                // Try finding closest if not exact match? 
                // For now just skip if not in range
                return;
            }
            xPos = xAxis.getPixelForValue(chart.data.labels[index]);
        }

        const ctx = chart.ctx;
        ctx.save();
        ctx.beginPath();
        ctx.moveTo(xPos, yAxis.top);
        ctx.lineTo(xPos, yAxis.bottom);
        ctx.lineWidth = 2;
        ctx.strokeStyle = '#ef4444'; // Red
        ctx.setLineDash([5, 5]);
        ctx.stroke();
        
        // Label
        ctx.fillStyle = '#ef4444';
        ctx.font = 'bold 12px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('LOG EVENT', xPos, yAxis.top - 5);
        
        ctx.restore();
    }
};

window.loadMetricData = async function() {
    const selector = document.getElementById('metric-selector');
    const metricName = selector.value;
    const rateToggle = document.getElementById('rate-toggle');
    const asRate = rateToggle ? rateToggle.checked : false;

    if (!metricName) return;

    const loading = document.getElementById('explorer-loading');
    const plotButton = document.getElementById('plot-button');
    const plotSpinner = document.getElementById('plot-spinner');
    
    if (loading) loading.classList.add('htmx-request');
    if (plotButton) plotButton.disabled = true;
    if (plotSpinner) plotSpinner.classList.add('htmx-request');

    try {
        const response = await fetch(`/api/metrics/data?metric=${encodeURIComponent(metricName)}`);
        const data = await response.json();

        window.renderExplorerChart(metricName, data, asRate);
    } catch (error) {
        console.error('Error fetching metric data:', error);
        alert('Failed to load metric data.');
    } finally {
        if (loading) loading.classList.remove('htmx-request');
        if (plotButton) plotButton.disabled = false;
        if (plotSpinner) plotSpinner.classList.remove('htmx-request');
    }
};

window.renderExplorerChart = function(title, seriesData, asRate) {
    const canvas = document.getElementById('explorerChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    
    const existingChart = Chart.getChart(canvas);
    if (existingChart) {
        existingChart.destroy();
    }

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || 
                       document.documentElement.getAttribute('data-theme') === 'ultra-dark' ||
                       document.documentElement.getAttribute('data-theme') === 'powershell';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    const colors = [
        '#36a2eb', '#ff6384', '#4bc0c0', '#ff9f40', '#9966ff', '#ffcd56', '#c9cbcf',
        '#00ff00', '#00ffff', '#ff00ff', '#ffff00'
    ];

    const targetT = getTargetTimestamp();

    const datasets = seriesData.map((s, index) => {
        const color = colors[index % colors.length];
        
        let processedData = [];
        
        // We map data to {x, y} where x is a timestamp number (ms).
        // This is required for 'parsing: false' + 'decimation'.
        if (asRate && s.data.length > 1) {
            for (let i = 1; i < s.data.length; i++) {
                const p1 = s.data[i-1];
                const p2 = s.data[i];
                
                const t1 = new Date(p1.x).getTime();
                const t2 = new Date(p2.x).getTime();
                const deltaSeconds = (t2 - t1) / 1000;
                
                if (deltaSeconds > 0) {
                    const rate = (p2.y - p1.y) / deltaSeconds;
                    if (rate >= 0) {
                        processedData.push({
                            x: t2, 
                            y: rate
                        });
                    }
                }
            }
        } else {
            processedData = s.data.map(p => {
                return {
                    x: new Date(p.x).getTime(),
                    y: p.y
                };
            });
        }

        return {
            label: s.label + (asRate ? ' (rate/s)' : ''),
            data: processedData,
            borderColor: color,
            backgroundColor: color,
            borderWidth: 2,
            fill: false,
            pointRadius: 0, // Disable points for performance by default
            tension: 0.1
        };
    });

    new Chart(ctx, {
        type: 'line',
        data: {
            datasets: datasets
        },
        plugins: [verticalLinePlugin],
        options: {
            responsive: true,
            maintainAspectRatio: false,
            parsing: false, // Critical for performance
            normalized: true, // Data is sorted
            animation: false, // Disable animation for large datasets
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            },
            plugins: {
                verticalLine: {
                    timestamp: targetT
                },
                decimation: {
                    enabled: true,
                    algorithm: 'lttb',
                    samples: 500, // Max pixels width-wise roughly
                    threshold: 500, // Only decimate if more than this many points
                },
                title: {
                    display: true,
                    text: title + (asRate ? ' (Rate)' : ''),
                    color: textColor
                },
                legend: {
                    labels: { color: textColor }
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    callbacks: {
                        title: function(context) {
                            if (context.length > 0) {
                                const d = new Date(context[0].parsed.x);
                                return d.toISOString().substring(11, 19);
                            }
                            return '';
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'linear', // Use linear scale for numeric timestamps
                    grid: { color: gridColor },
                    ticks: { 
                        color: textColor,
                        callback: function(value) {
                            return new Date(value).toISOString().substring(11, 19);
                        }
                    },
                    title: { display: true, text: 'Time (HH:mm:ss)', color: textColor }
                },
                y: {
                    beginAtZero: true,
                    grid: { color: gridColor },
                    ticks: { color: textColor },
                    title: { display: true, text: asRate ? 'Rate / s' : 'Value', color: textColor }
                }
            }
        }
    });
};

window.initSarChart = function(labels, userCpu, sysCpu, idleCpu) {
    const canvas = document.getElementById('sarCpuChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    const targetT = getTargetTimestamp();

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: '% User',
                    data: userCpu,
                    borderColor: 'rgba(54, 162, 235, 1)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: '% System',
                    data: sysCpu,
                    borderColor: 'rgba(255, 99, 132, 1)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: '% Idle',
                    data: idleCpu,
                    borderColor: 'rgba(75, 192, 192, 1)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    fill: false,
                    hidden: true,
                    pointRadius: 0
                }
            ]
        },
        plugins: [verticalLinePlugin],
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            plugins: {
                verticalLine: {
                    timestamp: targetT
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    title: { display: true, text: 'Percentage' },
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                },
                x: {
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                }
            }
        }
    });
};

window.initSarNetChart = function(labels, retrans, throughput) {
    const canvas = document.getElementById('sarNetChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    const targetT = getTargetTimestamp();

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'TCP Retransmissions (/s)',
                    data: retrans,
                    borderColor: 'rgba(255, 159, 64, 1)',
                    backgroundColor: 'rgba(255, 159, 64, 0.2)',
                    yAxisID: 'y',
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: 'Peak Throughput (MB/s)',
                    data: throughput,
                    borderColor: 'rgba(153, 102, 255, 1)',
                    backgroundColor: 'rgba(153, 102, 255, 0.2)',
                    yAxisID: 'y1',
                    fill: false,
                    pointRadius: 0
                }
            ]
        },
        plugins: [verticalLinePlugin],
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            plugins: {
                verticalLine: {
                    timestamp: targetT
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    display: true,
                    position: 'left',
                    beginAtZero: true,
                    title: { display: true, text: 'Retransmissions /s', color: 'rgba(255, 159, 64, 1)' },
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                },
                y1: {
                    type: 'linear',
                    display: true,
                    position: 'right',
                    beginAtZero: true,
                    title: { display: true, text: 'Throughput (MB/s)', color: 'rgba(153, 102, 255, 1)' },
                    grid: { drawOnChartArea: false },
                    ticks: { color: textColor }
                },
                x: {
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                }
            }
        }
    });
};

window.initShardCpuChart = function(labels, data) {
    const canvas = document.getElementById('shardCpuChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [{
                label: 'CPU Usage (%)',
                data: data,
                backgroundColor: 'rgba(54, 162, 235, 0.6)',
                borderColor: 'rgba(54, 162, 235, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    grid: {
                        color: gridColor
                    },
                    ticks: {
                        color: textColor
                    }
                },
                x: {
                    grid: {
                        color: gridColor
                    },
                    ticks: {
                        color: textColor
                    }
                }
            }
        }
    });
};

// Initialize correlation features
document.addEventListener('DOMContentLoaded', () => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('t')) {
        // Scroll to the historical charts (SAR) which are most likely to show the event
        const historicalChart = document.getElementById('sarCpuChart');
        if (historicalChart) {
            setTimeout(() => {
                historicalChart.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }, 800);
        }
    }
});
