document.body.addEventListener('htmx:responseError', function(evt) {
    const content = document.getElementById("analysis-content");
    if (content) {
        content.innerHTML = 
            '<div style="color: red; text-align: center; padding: 20px;">' +
            '<h3>Error Loading Analysis</h3>' +
            '<p>Server returned status: ' + evt.detail.xhr.status + '</p>' +
            '<p>Check server logs for details.</p>' +
            '</div>';
    }
});

document.body.addEventListener('htmx:sendError', function(evt) {
    const content = document.getElementById("analysis-content");
    if (content) {
        content.innerHTML = 
            '<div style="color: red; text-align: center; padding: 20px;">' +
            '<h3>Connection Error</h3>' +
            '<p>Failed to connect to server. Is it running?</p>' +
            '</div>';
    }
});
