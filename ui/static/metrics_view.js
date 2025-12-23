// metrics_view.js - Global chart initialization functions

window.explorerChart = null;

window.loadMetricData = async function() {
    const selector = document.getElementById('metric-selector');
    const metricName = selector.value;
    const rateToggle = document.getElementById('rate-toggle');
    const asRate = rateToggle ? rateToggle.checked : false;

    if (!metricName) return;

    const loading = document.getElementById('explorer-loading');
    if (loading) loading.style.display = 'flex';

    try {
        const response = await fetch(`/api/metrics/data?metric=${encodeURIComponent(metricName)}`);
        const data = await response.json();

        window.renderExplorerChart(metricName, data, asRate);
    } catch (error) {
        console.error('Error fetching metric data:', error);
        alert('Failed to load metric data.');
    } finally {
        if (loading) loading.style.display = 'none';
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
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
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
